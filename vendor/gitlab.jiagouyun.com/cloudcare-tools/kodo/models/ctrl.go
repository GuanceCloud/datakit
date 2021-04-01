package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/utils"
)

var (
	MaxNote = 30

	// seconds
	MaxUploaderInterval = int64(15 * 60)
	MinUploaderInterval = int64(15)
)

func getUserPwd(auth, rw string) (user, pwd string, err error) {
	var authMap map[string]InfluxAuth
	err = json.Unmarshal([]byte(auth), &authMap)
	if err != nil {
		l.Errorf("parse json %s failed: %s", auth, err.Error())
		return "", "", utils.ErrInvalidJson
	}

	var password string
	if authMap[rw].Password != "" {
		password = authMap[rw].Password
	} else {
		pwd = authMap[rw].PwdEncrypt
		password = utils.DecipherByAES(pwd, config.C.Secret.EncryptKey)
	}
	return authMap[rw].UserName, password, nil
}

func CleanTokenCache(token string) error {
	_, err := config.Redis.Del("tkn_info:" + token).Result()
	return err
}

func CleanDBInfoCache(dbUUID string) error {
	_, err := config.Redis.Del(`dbinfo:` + dbUUID).Result()
	return err
}

func SetDropMetrics(dbUUID, measurement string) {
	config.Redis.SAdd(`dropping_metrics:`+dbUUID, measurement)
	config.Redis.Expire(`dropping_metrics:`+dbUUID, 2*time.Hour)
}

func DelDropMetrics(dbUUID, measurement string) {
	config.Redis.SRem(`dropping_metrics:`+dbUUID, measurement)

	count, err := config.Redis.SCard(`dropping_metrics:` + dbUUID).Result()
	if err != nil {
		return
	}

	if count == 0 {
		config.Redis.Del(`dropping_metrics:` + dbUUID)
	}
}

func GetDroppingMetrics(dbUUID string) ([]string, error) {
	return config.Redis.SMembers(`dropping_metrics:` + dbUUID).Result()
}

func GetTokenInfo(token string) (*Token, error) {
	var uuid, dbuuid string
	var versionType, billState string

	key := "tkn_info:" + token

	tkninfo, err := config.Redis.HGetAll(key).Result()
	if err == nil && len(tkninfo) > 0 && tkninfo["wsuuid"] != `` {

		l.Debugf("tkninfo: %+#v", tkninfo)
		return &Token{
			WsUUID: tkninfo["wsuuid"],
			DBUUID: tkninfo["dbuuid"],
			Token:  token,

			BillState: tkninfo[utils.BillState],
			VerType:   tkninfo[utils.VerType],
		}, nil

	}

	if err != nil {
		l.Warnf("[warn] %s, ignored, get token:%s info from DB", err.Error(), token)
	}

	//TODO  billstatus
	err = Stmts[`qToken`].QueryRow(StatusOK, token).Scan(&uuid, &dbuuid, &versionType, &billState)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrTokenNotFound
		} else {
			l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
			return nil, err
		}
	}

	//TODO billstatus
	tkn := &Token{
		WsUUID:    uuid,
		DBUUID:    dbuuid,
		Token:     token,
		VerType:   versionType,
		BillState: billState,
	}

	config.Redis.HMSet(key, map[string]interface{}{
		"wsuuid":        uuid,
		"dbuuid":        dbuuid,
		utils.BillState: billState,
		utils.VerType:   versionType,
	})

	config.Redis.Expire(key, 24*time.Hour)

	return tkn, nil
}

func UpdateAgent(dw *Agent) error {
	key := "dataway_ver:" + dw.UUID
	oldver, err := config.Redis.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			l.Debugf("key %s not found", key)
			goto __writeDB
		} else {
			l.Errorf("%s", err.Error())
			return err
		}
	}

	if oldver == dw.Version {
		return nil
	}
	// else: go down

__writeDB:

	l.Debugf("dataway version(%s) %s <> %s", key, oldver, dw.Version)

	// version changed, update new version into database
	if _, err := Stmts[`uAgent`].Exec(dw.Version, StatusOK, dw.UploadAt, dw.UUID); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}

	config.Redis.Set(key, dw.Version, 0)
	return nil
}

func CreateInfluxDB(idb *InfluxDBInfo) (string, error) {
	now := time.Now().Unix()
	if _, err := Stmts[`iInfluxdb`].Exec(idb.DB, idb.CQRP, idb.RP, idb.UUID, idb.RPUUID, idb.InfluxInstanceUUID, now, now); err != nil {
		l.Errorf("%s, ignored", err)
		return "", err
	}
	return idb.UUID, nil
}

func UpdateInfluxDB(idb *InfluxDBInfo) error {
	now := time.Now().Unix()
	if _, err := Stmts[`uInfluxdb`].Exec(idb.RP, now, idb.UUID); err != nil {
		l.Errorf("%s, ignored", err)
		return err
	}
	return nil

}

func QueryInfluxInfoReadOnly(dbUUID string) (ii *InfluxDBInfo, err error) {
	return queryInfluxInfo(dbUUID, utils.InfluxPermReadOnly)
}

func QueryInfluxInfoReadWrite(dbUUID string) (ii *InfluxDBInfo, err error) {
	return queryInfluxInfo(dbUUID, utils.InfluxPermReadWrite)
}

func QueryInfluxInfoAdmin(dbUUID string) (ii *InfluxDBInfo, err error) {
	return queryInfluxInfo(dbUUID, utils.InfluxPermAdmin)
}

func QueryWorkspaceInfoReadWrite(wsid string) (ii *InfluxDBInfo, err error) {
	return queryWorkspaceInfluxInfo(wsid, utils.InfluxPermReadWrite)
}

func QueryWorkspaceInfoReadOnly(wsid string) (ii *InfluxDBInfo, err error) {
	return queryWorkspaceInfluxInfo(wsid, utils.InfluxPermReadOnly)
}

func queryWorkspaceInfluxInfo(wsid string, rw string) (ii *InfluxDBInfo, err error) {
	var dbname, instid, dbhost, auth string
	err = Stmts[`qWorkspaceDBInfo`].QueryRow(wsid).Scan(&dbname, &instid, &dbhost, &auth)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInfluxDBNotFound
		} else {
			l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
			return
		}
	}

	user, pwd, err := getUserPwd(auth, rw)
	if err != nil {
		return nil, err
	}

	return &InfluxDBInfo{
		DB:                 dbname,
		Host:               dbhost,
		User:               user,
		InfluxInstanceUUID: instid,
		Pwd:                pwd,
	}, nil
}

func queryInfluxInfo(dbUUID string, rw string) (ii *InfluxDBInfo, err error) {

	//a.db, b.host, b.user, b.pwd, b.uuid, a.uuid
	var host, isuuid, rp, db, cqrp string
	var auth string

	err = Stmts[`qInfluxdb`].QueryRow(dbUUID, StatusOK).Scan(&host, &auth, &isuuid, &rp, &db, &cqrp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInfluxDBNotFound
		} else {
			l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
			return
		}
	} else {
		user, pwd, err := getUserPwd(auth, rw)
		if err != nil {
			return nil, err
		}

		return &InfluxDBInfo{
			DB:                 db,
			Host:               host,
			User:               user,
			Pwd:                pwd,
			UUID:               dbUUID,
			InfluxInstanceUUID: isuuid,
			RP:                 rp,
			CQRP:               cqrp,
		}, nil
	}
}

func QueryInfluxInstanceByUUID(uuid, rw string) (*InfluxInstance, error) {
	var status, dbcount int
	var host, user, pwd, auth string

	err := Stmts[`qInfluxInstanceByUUID`].QueryRow(StatusOK, uuid).Scan(&host, &auth, &dbcount, &status, &uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInfluxNotFound
		} else {
			l.Errorf(" DB Error %s, %+#v", err.Error(), DB.Stats())
			return nil, err
		}
	}

	user, pwd, err = getUserPwd(auth, rw)
	if err != nil {
		return nil, err
	}

	return &InfluxInstance{
		Host:          host,
		User:          user,
		UUID:          uuid,
		Pwd:           pwd,
		Authorization: auth,
		DBCount:       dbcount,
		Status:        status,
	}, nil
}

func QueryAllInfluxInstanceAdmin() (res []*InfluxInstance, err error) {
	return queryAllInfluxInstance(utils.InfluxPermAdmin)
}

func QueryAllInfluxInstanceReadOnly() (res []*InfluxInstance, err error) {
	return queryAllInfluxInstance(utils.InfluxPermReadWrite)
}

func QueryAllInfluxInstanceReadWrite() (res []*InfluxInstance, err error) {
	return queryAllInfluxInstance(utils.InfluxPermReadWrite)
}

func queryAllInfluxInstance(rw string) (res []*InfluxInstance, err error) {
	var rows *sql.Rows

	rows, err = Stmts[`qInfluxInstances`].Query(StatusOK)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no influx instance available")
		} else {
			return nil, err
		}
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	for rows.Next() {
		//host, user, pwd, dbcount, status, uuid
		var status, dbcount int
		var uuid, host, user, pwd, auth string

		if err = rows.Scan(&host, &auth, &dbcount, &status, &uuid); err != nil {
			l.Errorf("%s, ignored", err.Error())
			return nil, err
		}

		user, pwd, err := getUserPwd(auth, rw)
		if err != nil {
			return nil, err
		}

		res = append(res, &InfluxInstance{
			Host:          host,
			User:          user,
			UUID:          uuid,
			Pwd:           pwd,
			Authorization: auth,
			DBCount:       dbcount,
			Status:        status,
		})
	}

	return res, nil
}

func ModifyInfluxInstance(sr *InfluxInstance) error {
	now := time.Now().Unix()
	if _, err := Stmts[`uInfluxInstance`].Exec(sr.Host, sr.Authorization, sr.Status, now, sr.UUID); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func GetInstDBs(instUUID string) ([]*InfluxDBInfo, error) {
	rows, err := Stmts["getInstDBs"].Query(instUUID, StatusOK)
	if err != nil {
		l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	res := []*InfluxDBInfo{}

	for rows.Next() {
		var ifdb InfluxDBInfo
		if err := rows.Scan(
			&ifdb.UUID,
			&ifdb.DB,
		); err != nil {
			l.Errorf("%s", err.Error())
			return nil, err
		}

		res = append(res, &ifdb)
	}

	return res, nil
}

func GetWsDBUIDs() (map[string]string, error) {
	rows, err := Stmts["qWorkSpacesDBUUIDs"].Query(StatusOK)
	if err != nil {
		l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	res := map[string]string{}

	for rows.Next() {
		var wsUUID, dbUUID string
		if err := rows.Scan(
			&wsUUID,
			&dbUUID,
		); err != nil {
			l.Errorf("%s", err.Error())
			return nil, err
		}
		res[dbUUID] = wsUUID
	}

	return res, nil

}

func GetWsDBUID(wsUUID string) (string, error) {
	dbUUID := ``

	err := Stmts["qWorkSpacesDBUUID"].QueryRow(wsUUID).Scan(&dbUUID)
	if err != nil {

		if err == sql.ErrNoRows {
			return "", utils.ErrInfluxDBNotFound
		} else {
			l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
			return "", err
		}

	}

	return dbUUID, nil

}

func QueryInfluxCQ(uuid, workspaceUUID string) (*InfluxCQ, error) {
	var dbUUID, sampleEvery, sampleFor, measurement, rp, cqrp, aggrFunc, groupByTime, groupByOffset string

	err := Stmts[`qInfluxCQ`].QueryRow(uuid, workspaceUUID, StatusOK).Scan(&dbUUID,
		&sampleEvery, &sampleFor, &measurement, &rp, &cqrp, &aggrFunc, &groupByTime, &groupByOffset)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInfluxCQNotFound
		} else {
			l.Errorf(" DB Error %s, %+#v", err.Error(), DB.Stats())
			return nil, err
		}
	}

	return &InfluxCQ{
		InfluxdbUUID:  dbUUID,
		SampleEvery:   sampleEvery,
		SampleFor:     sampleFor,
		Measurement:   measurement,
		CQUUID:        uuid,
		WorkspaceUUID: workspaceUUID,
		RP:            rp,
		CQRP:          cqrp,
		FuncAggr:      aggrFunc,
		GroupByOffset: groupByOffset,
		GroupByTime:   groupByTime,
	}, nil

}

func ModifyInfluxCQ(cq *InfluxCQ) error {
	now := time.Now().Unix()

	if _, err := Stmts[`uInfluxCQ`].Exec(cq.SampleEvery, now, cq.GroupByTime, cq.FuncAggr, cq.CQUUID, cq.WorkspaceUUID); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func CreateInfluxCQ(cq *InfluxCQ) (string, error) {
	now := time.Now().Unix()
	// uuid, dbUUID, workspaceUUID, sampleEvery, sampleFor, measurement,
	// 				rp,cqrp,aggrFunc,groupByTime,groupByOffset,createAt
	if _, err := Stmts[`iInfluxCQ`].Exec(cq.CQUUID, cq.InfluxdbUUID, cq.WorkspaceUUID, cq.SampleEvery,
		cq.SampleFor, cq.Measurement, cq.RP, cq.CQRP, cq.FuncAggr, cq.GroupByTime, cq.GroupByOffset, now); err != nil {
		l.Errorf("%s, ignored", err)
		return "", err
	}
	return cq.CQUUID, nil
}

func DropInfluxCQ(uuid, workspaceUUID string) error {
	if _, err := Stmts[`dInfluxCQ`].Exec(uuid, workspaceUUID); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func DropInfluxCQByUUID(uuid string) error {
	if _, err := Stmts[`dInfluxCQByuuid`].Exec(uuid); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func QueryInfluxCQByM(dbUUID, measurement string) (*InfluxCQ, error) {
	var uuid, sampleEvery, workspaceUUID, sampleFor, rp, cqrp, aggrFunc, groupByTime, groupByOffset string

	err := Stmts[`qInfluxCQByM`].QueryRow(dbUUID, measurement, StatusOK).Scan(&uuid, &workspaceUUID,
		&sampleEvery, &sampleFor, &measurement, &rp, &cqrp, &aggrFunc, &groupByTime, &groupByOffset)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInfluxCQNotFound
		} else {
			l.Errorf(" DB Error %s, %+#v", err.Error(), DB.Stats())
			return nil, err
		}
	}

	return &InfluxCQ{
		InfluxdbUUID:  dbUUID,
		SampleEvery:   sampleEvery,
		SampleFor:     sampleFor,
		Measurement:   measurement,
		CQUUID:        uuid,
		WorkspaceUUID: workspaceUUID,
		RP:            rp,
		CQRP:          cqrp,
		FuncAggr:      aggrFunc,
		GroupByOffset: groupByOffset,
		GroupByTime:   groupByTime,
	}, nil

}

func QueryInfluxCQsByWsUUID(wsUUID string) (res []*InfluxCQ, err error) {
	var rows *sql.Rows

	rows, err = Stmts[`qInfluxCQsByWsUUID`].Query(wsUUID, StatusOK)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no any influx-cq")
		} else {
			return nil, err
		}
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	for rows.Next() {
		//host, user, pwd, dbcount, status, uuid
		var uuid, dbUUID, sampleEvery, sampleFor, measurement, rp, cqrp, aggrFunc, groupByTime, groupByOffset string

		if err = rows.Scan(&uuid, &dbUUID, &sampleEvery, &sampleFor, &measurement, &rp, &cqrp,
			&aggrFunc, &groupByTime, &groupByOffset); err != nil {
			l.Errorf("%s, ignored", err.Error())
			return nil, err
		}

		res = append(res, &InfluxCQ{
			InfluxdbUUID:  dbUUID,
			SampleEvery:   sampleEvery,
			SampleFor:     sampleFor,
			Measurement:   measurement,
			CQUUID:        uuid,
			WorkspaceUUID: wsUUID,
			RP:            rp,
			CQRP:          cqrp,
			FuncAggr:      aggrFunc,
			GroupByOffset: groupByOffset,
			GroupByTime:   groupByTime,
		})
	}

	return res, nil
}

func QueryConfig() (res map[string]string, err error) {
	//
	var rows *sql.Rows

	res = map[string]string{}
	rows, err = Stmts[`qConfig`].Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no any main_config")
		} else {
			return nil, err
		}
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	for rows.Next() {

		var keyCode, value string

		if err = rows.Scan(&keyCode, &value); err != nil {
			l.Errorf("%s, ignored", err.Error())
			return nil, err
		}

		res[keyCode] = value
	}

	return res, nil
}

func GetObjectsInfo(source string) (string, error) {
	key := `dfobjects_info`

	contentmd5, err := config.Redis.HGet(key, source).Result()
	if err != nil {
		l.Warnf(" DB Error %s, %+#v", err.Error(), DB.Stats())
		return "", err
	}

	return contentmd5, nil
}

func SetObjectsInfo(source string, value interface{}) error {
	_, err := config.Redis.HSet(`dfobjects_info`, source, value).Result()
	if err != nil {
		l.Errorf("%s", err.Error())
		return err
	}
	return nil
}

func GetObjectClassTagsFields(wsUUID, className string) (res *ObjectClassCfg, err error) {
	var tags, fields string

	err = Stmts[`qObjectClass`].QueryRow(wsUUID, className, StatusOK).Scan(&tags, &fields)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrInfluxCQNotFound
		} else {
			l.Errorf(" DB Error %s, %+#v", err.Error(), DB.Stats())
			return nil, err
		}
	}

	return &ObjectClassCfg{
		WsUUID:    wsUUID,
		ClassName: className,
		Tags:      tags,
		Fields:    fields,
	}, nil
}

func GetObjectClassTagsFieldsByUUID(wsUUID string) (res map[string]*ObjectClassCfg, err error) {

	var rows *sql.Rows
	res = map[string]*ObjectClassCfg{}
	rows, err = Stmts[`qObjectClassByUUID`].Query(wsUUID, StatusOK)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNoObjectClassCfg
		} else {
			return nil, err
		}
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	for rows.Next() {

		var tags, fields, className string
		if err = rows.Scan(&tags, &fields, &className); err != nil {
			l.Errorf("%s, ignored", err.Error())
			return nil, err
		}

		res[wsUUID+className] = &ObjectClassCfg{
			WsUUID:    wsUUID,
			ClassName: className,
			Tags:      tags,
			Fields:    fields,
		}

	}

	return res, nil
}

func UpdateObjectClassTagsFields(oc *ObjectClassCfg) error {
	now := time.Now().Unix()

	if _, err := Stmts[`uObjectClass`].Exec(oc.Tags, oc.Fields, now, oc.WsUUID, oc.ClassName); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func AddObjectClassTagsFields(oc *ObjectClassCfg) error {

	now := time.Now().Unix()
	if _, err := Stmts[`iObjectClass`].Exec(oc.UUID, oc.WsUUID, oc.ClassName, oc.Tags, oc.Fields, now); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func GetLogExtractUrl(wsUUID, source string) (string, error) {
	//
	key := `logExRule:` + wsUUID

	url, err := config.Redis.HGet(key, source).Result()
	if err != nil {
		//不存在更新redis
		res, err := QueryLogExtractRule(wsUUID)
		if err != nil {
			l.Errorf("%s", err.Error())
			return "", err
		}

		fs := map[string]interface{}{}
		for _, r := range res {

			fs[r.Source] = r.Url
			if r.Source == source {
				url = r.Url
			}
		}

		config.Redis.HMSet(key, fs)
		config.Redis.Expire(key, 24*time.Hour)
	}

	return url, nil
}

// log_extract_rule 查找
func QueryLogExtractRule(wsUUID string) (res []*LogExtractRule, err error) {
	var rows *sql.Rows

	rows, err = Stmts[`qLogExtractRule`].Query(wsUUID, StatusOK)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("no any influx-cq")
		} else {
			return nil, err
		}
	}

	if rows == nil {
		return nil, nil
	}

	defer rows.Close()

	for rows.Next() {
		//workspaceUUID, source, url
		var source, url, wsUUID string

		if err = rows.Scan(&url, &wsUUID, &source); err != nil {
			l.Errorf("%s, ignored", err.Error())
			return nil, err
		}

		res = append(res, &LogExtractRule{
			WsUUID: wsUUID,
			Source: source,
			Url:    url,
		})
	}

	return res, nil

}

func SetLogExtractUrl(wsUUID, source, url string) error {
	key := `logExRule:` + wsUUID

	if len(url) == 0 {
		_, err := config.Redis.HDel(key, source).Result()
		if err != nil {
			l.Errorf("%s", err.Error())
			return err
		}

	} else {
		_, err := config.Redis.HSet(key, source, url).Result()
		if err != nil {
			l.Errorf("%s", err.Error())
			return err
		}
	}

	return nil
}

func UpdateDKStatus(dkUUID string, heartbeat int64, status int) error {
	now := time.Now().Unix()
	_, err := Stmts[`updateDKStatus`].Exec(heartbeat, status, now, dkUUID)
	return err
}

func CheckCliToken(wsid, token string) error {

	var cliToken string

	err := Stmts[`qWorkspaceCliToken`].QueryRow(wsid).Scan(&cliToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrWorkspaceNotFound
		} else {
			l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
			return err
		}
	}

	if cliToken == token {
		return nil
	}

	return utils.ErrInvalidCliToken
}

func GetMigInfluxDBs(wsuids string) ([]*InfluxDBInfo, error) {

	rows, err := Stmts["qMigInfluxDBInfo"].Query(wsuids)
	if wsuids == `` {
		rows, err = Stmts["qAllInfluxDBInfo"].Query(StatusOK)
	}
	if err != nil {
		l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	res := []*InfluxDBInfo{}

	var durationSet string
	for rows.Next() {
		var ifdb InfluxDBInfo
		if err := rows.Scan(&ifdb.DB, &ifdb.CQRP, &ifdb.RP, &durationSet); err != nil {
			l.Errorf("%s", err.Error())
			return nil, err
		}

		dm := map[string]interface{}{}
		err = json.Unmarshal([]byte(durationSet), &dm)
		if err != nil {
			l.Errorf("parse json %s failed: %s", durationSet, err.Error())
			return nil, utils.ErrInvalidJson
		}

		duration, ok := dm[`rp`]
		if !ok {
			l.Errorf("parse json %s failed: %s", durationSet, err.Error())
			return nil, utils.ErrInvalidJson
		}

		ifdb.Duration = duration.(string)

		res = append(res, &ifdb)
	}

	return res, nil
}

func DropTask(ak string) error {
	if _, err := Stmts[`dDropTasks`].Exec(ak); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func BatchAddDialTasks(tasks []dt.Task) error {

	sqlstr := "INSERT INTO task(uuid, external_id, region, class, task, createAt, updateAt) VALUES "
	extras := []string{}
	values := []interface{}{}

	for _, t := range tasks {

		j, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}

		values = append(values,
			cliutils.XID("dialt_"),
			t.ID(),
			t.RegionName(),
			t.Class(),
			string(j),
			t.UpdateTimeUs(),
			t.UpdateTimeUs())
		extras = append(extras, "(?,?,?,?,?,?,?)")
	}

	sqlstr += strings.Join(extras, ",")

	stmt, err := prepare(DB, sqlstr)
	if err != nil {
		return err
	}

	defer stmt.Close()

	if _, err := stmt.Exec(values...); err != nil {
		switch me := err.(type) {
		case *mysql.MySQLError:
			l.Warnf("mysql error: %+#v", me)
			return err

		default:
			l.Error(err)
			return err
		}
	}

	return nil
}

func GetKeyConfig(wsuid, keycode string) (*KeyConfig, error) {

	var value, uuid string
	err := Stmts[`qKeyConfig`].QueryRow(wsuid, keycode, StatusOK).Scan(&uuid, &value)
	if err != nil {
		if err == sql.ErrNoRows {
			l.Errorf(`%s`, err.Error())
			return nil, utils.ErrKeyConfigNotFound
		} else {
			l.Errorf("DB Error %s, %+#v", err.Error(), DB.Stats())
			return nil, err
		}
	}

	return &KeyConfig{
		UUID:    uuid,
		WsUUID:  wsuid,
		KeyCode: keycode,
		Value:   utils.DecipherByAES(value, config.C.Secret.EncryptKey),
	}, nil

}

func GetKeyConfigs(wsuid string, keycodes []string) ([]*KeyConfig, error) {
	kcs := []string{}
	kcvs := []interface{}{}
	kcvs = append(kcvs, wsuid)
	kcvs = append(kcvs, StatusOK)
	for _, kc := range keycodes {
		kcs = append(kcs, `?`)
		kcvs = append(kcvs, kc)
	}

	sqlstr := `SELECT uuid, value, keycode FROM main_key_config WHERE workspaceUUID=? AND status=? AND keyCode IN (` + strings.Join(kcs, `,`) + `)`
	stmt, err := DB.Prepare(sqlstr)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	defer stmt.Close()

	res := []*KeyConfig{}
	rows, err := stmt.Query(kcvs...)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	for rows.Next() {
		var value, uuid, keycode string
		if err := rows.Scan(&uuid, &value, &keycode); err != nil {
			l.Errorf("%s", err.Error())
			return nil, err
		}

		res = append(res, &KeyConfig{
			UUID:    uuid,
			WsUUID:  wsuid,
			KeyCode: keycode,
			Value:   utils.DecipherByAES(value, config.C.Secret.EncryptKey),
		})
	}

	return res, nil

}

func UpdateKeyConfigValue(kc KeyConfig) error {

	value := utils.CipherByAES(kc.Value, config.C.Secret.EncryptKey)
	now := time.Now().Unix()
	if _, err := Stmts[`uKeyConfig`].Exec(value, now, kc.WsUUID, kc.KeyCode); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func PullDialTask(region string, sinceUs int64) (tasks map[string][]string, err error) {

	rows, err := Stmts[`qPullDialTestingTasks`].Query(region, sinceUs)
	if err != nil {
		l.Error(err)
		return
	}

	tasks = map[string][]string{}

	exids := map[string]bool{}

	for rows.Next() {
		var task, class, external_id string
		if err = rows.Scan(&task, &class, &external_id); err != nil {
			return
		} else {
			if _, ok := exids[external_id]; ok {
				continue
			}

			switch class {
			case dt.ClassHTTP:
				exids[external_id] = true
				tasks["HTTP"] = append(tasks["HTTP"], task)

			case dt.ClassTCP:
				// TODO
			case dt.ClassDNS:
				// TODO
			default:
				// TODO
			}
		}
	}

	return
}

func GetDialTestingAKInfo(ak string) (*DialTestingAkInfo, error) {
	var owner, sk string
	if err := Stmts[`qDialTestingAK`].QueryRow(ak).Scan(&owner, &sk); err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrAKNotFound
		}
		return nil, err
	}

	return &DialTestingAkInfo{
		Owner: owner,
		AK:    ak,
		SK:    sk,
	}, nil
}

func CreateDialtestingAK(owner string) (*DialTestingAkInfo, error) {
	ak := cliutils.CreateRandomString(20)
	sk := cliutils.CreateRandomString(40)
	akUUID := cliutils.XID(`ak_`)

	now := time.Now().Unix()

	if _, err := Stmts[`iDialTestingAKSK`].Exec(akUUID, ak, sk, owner, "OK", 0, now, now); err != nil {
		l.Errorf("%s", err.Error())
		return nil, err
	}

	return &DialTestingAkInfo{
		AK: ak, SK: sk, Owner: owner,
	}, nil
}

func NewKeyConfig(kc KeyConfig) error {

	value := utils.CipherByAES(kc.Value, config.C.Secret.EncryptKey)
	now := time.Now().Unix()

	if _, err := Stmts[`iKeyConfig`].Exec(kc.UUID, kc.WsUUID, kc.KeyCode, value, kc.Desp, StatusOK, now); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}

func DeleteKeyConfig(kc KeyConfig) error {

	now := time.Now().Unix()
	if _, err := Stmts[`dKeyConfig`].Exec(kc.Status, now, kc.WsUUID, kc.KeyCode); err != nil {
		l.Errorf("%s, ignored", err.Error())
		return err
	}
	return nil
}
