//go:build linux
// +build linux

package protodec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

// PgsqlProtoDetect 只解析 Query 之类的请求 因此firstDstPort就是Pgsql的端口.
func PgsqlProtoDetect(data []byte, actSize int) (L7Protocol, ProtoDecPipe, bool) {
	if checkPgsql(data) {
		return ProtoPgsql, newPgsqlDecPipe(ProtoPgsql), true
	}
	return ProtoUnknown, nil, false
}

func checkPgsql(payload []byte) bool {
	inf := &pgsqlInfo{}
	inf.direction = clientToServer
	return inf.parse(payload) == nil
}

type pgsqlInfo struct {
	meta ProtoMeta

	haveParsed    bool
	readyToExport bool

	resource     string
	resourceType string
	comment      string
	affectedRow  int
	statusCode   string
	errMsg       string
	msg          string
	status       int

	reqType   byte
	respType  byte
	direction Direction

	reqBytes  int
	respBytes int

	ktime [4]uint64
	ts    int64
	dur   [2]uint64
}

func (p *pgsqlInfo) parse(payload []byte) error {
	if len(payload) == 0 {
		return errors.New("payload empty")
	}
	offset := 0

	for offset < len(payload) {
		tag, length, err := readPgsqlBlock(payload[offset:])
		if err != nil {
			break
		}
		var ok bool
		switch p.direction { //nolint:exhaustive
		case clientToServer:
			if ok, err = p.parseRequest(payload[offset+5:offset+5+length], tag); err != nil {
				return err
			}
		case serverToClient:
			if ok, err = p.parseResponse(payload[offset+5:offset+5+length], tag); err != nil {
				return err
			}
			if ok {
				p.readyToExport = true
			}
		default:
			return ErrUnknownProto
		}

		if ok && !p.haveParsed {
			p.haveParsed = true
		}
		offset += length + PgsqlTypeOffset + PgsqlTypeLenOffest
	}
	if p.haveParsed {
		return nil
	}
	return ErrUnknownProto
}

func (p *pgsqlInfo) parseRequest(payload []byte, tag byte) (bool, error) {
	switch tag {
	case 'Q': // simple query
		p.reqType = tag
		idx := bytes.IndexByte(payload, '\x00')
		if idx == len(payload)-1 {
			sql := payload[:idx]
			if err := p.populateInfo(sql); err != nil {
				return false, err
			}
		} else {
			return false, ErrUnknownProto
		}

	case 'P': // parse
		p.reqType = tag

		begin := bytes.IndexByte(payload, '\x00')
		if begin == -1 {
			return false, ErrUnknownProto
		}

		end := bytes.IndexByte(payload[begin+1:], '\x00')
		// end := sqlIter.nextIndex()
		if end == -1 {
			return false, ErrUnknownProto
		}
		endGlobal := begin + end + 1
		if endGlobal > len(payload) {
			return false, ErrUnknownProto
		}
		sql := payload[begin+1 : endGlobal]
		if err := p.populateInfo(sql); err != nil {
			return false, err
		}
	case 'B', 'F', 'C', 'D', 'H', 'S', 'X', 'd', 'c', 'f':
		return false, nil
	default:
		return false, ErrUnknownProto
	}
	return true, nil
}

func (p *pgsqlInfo) parseResponse(payload []byte, tag byte) (bool, error) {
	switch tag {
	case 'C':
		p.respType = tag

		idx := bytes.IndexByte(payload, '\x20')
		if idx == -1 {
			return false, ErrUnknownProto
		}
		op := string(bytes.ToUpper(payload[:idx]))
		switch op {
		case "INSERT":
			payload = payload[idx+1:]
			if idx = bytes.IndexByte(payload, '\x20'); idx == -1 {
				return false, ErrUnknownProto
			}
			payload = payload[idx+1:]
			end := bytes.IndexByte(payload, '\x00')
			if end == -1 {
				return false, ErrUnknownProto
			}
			rows := payload[:end]
			if rowsAffect, err := strconv.Atoi(string(rows)); err == nil {
				p.affectedRow = rowsAffect
			} else {
				return false, ErrUnknownProto
			}
		case "DELETE", "UPDATE", "SELECT", "MERGE", "MOVE", "FETCH", "COPY":
		default:
		}
	case 'E':
		p.respType = tag
		for i := 0; i < 2; i++ {
			idx := bytes.IndexByte(payload, '\x00')
			if idx == -1 {
				return false, ErrUnknownProto
			}
			payload = payload[idx+1:]
		}

		idx := bytes.IndexByte(payload, '\x00')
		if idx == -1 {
			return false, ErrUnknownProto
		}
		if payload[0] != 'C' {
			return false, ErrUnknownProto
		}
		p.statusCode = string(payload[1:idx])
		p.errMsg, p.status = code2Msg(p.statusCode)

		payload = payload[idx+1:]
		idx = bytes.IndexByte(payload, '\x00')
		if idx == -1 {
			return false, ErrUnknownProto
		}
		p.msg = string(payload[1:idx])
	case 'Z', 'I', '1', '2', '3', 'S', 'K', 'T', 'n', 'N', 't', 'D',
		'G', 'H', 'W', 'd', 'c':
		return false, nil
	default:
		return false, ErrUnknownProto
	}
	return true, nil
}

func (p *pgsqlInfo) populateInfo(sql []byte) error {
	comment, command, clean := trimCommentGetFirst(sql, 12)

	var resource []byte
	if ok := p.isPgsql(sql); ok {
		if output, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", string(clean)); err == nil && output != nil {
			o := []byte(output.Query)
			validLen := utf8ValidLength(o)
			resource = o[:validLen]
		} else {
			resource = []byte(string(bytes.Runes(clean)))
		}
	} else {
		return ErrUnknownProto
	}

	p.resource = string(bytes.Runes(resource))
	p.comment = strings.TrimSpace(string(bytes.Runes(comment)))
	p.resourceType = string(bytes.Runes(command))

	return nil
}

type pgsqlDecPipe struct {
	direction comm.Direcion

	infCache     []*pgsqlInfo
	inf          *pgsqlInfo
	reqResp      int
	firstDstPort uint32
	connClosed   bool
}

func (dec *pgsqlDecPipe) Decode(txRx comm.NICDirection, data *comm.NetwrkData,
	ts int64, thrTr threadTrace,
) {
	if len(data.Payload) == 0 {
		return
	}

	// 第一次进decode的一定为Request,此时firstDstPort为server port
	if dec.firstDstPort == 0 {
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			dec.firstDstPort = data.Conn.Sport
		case comm.NICDEgress:
			dec.firstDstPort = data.Conn.Dport
		}
	}

	if dec.inf == nil {
		dec.inf = &pgsqlInfo{}
	}

	direction := directionUnknown
	switch txRx { //nolint:exhaustive
	case comm.NICDIngress:
		direction = dec.determineDirection(data.Conn.Sport, dec.inf)
	case comm.NICDEgress:
		direction = dec.determineDirection(data.Conn.Dport, dec.inf)
	}

	dec.inf.direction = direction
	inf := dec.inf
	if err := inf.parse(data.Payload); err != nil {
		return
	}

	switch direction { //nolint:exhaustive
	case clientToServer:
		dec.reqResp = 1
		if dec.direction == comm.DUnknown {
			switch txRx { //nolint:exhaustive
			case comm.NICDIngress:
				dec.direction = comm.DIn
			case comm.NICDEgress:
				dec.direction = comm.DOut
			}
		}

		if dec.direction == comm.DIn {
			inf.meta.InnerID = thrTr.Insert(dec.direction, int32(data.Conn.Pid),
				data.Thread, data.TSTail)
		}

		inf.ts = ts
		inf.meta.Threads[0] = data.Thread
		inf.meta.ReqTCPSeq = data.TCPSeq
		inf.dur[0] = data.TS
		inf.ktime[0] = data.TSTail

	case serverToClient:
		dec.reqResp = 2

		inf.meta.Threads[1] = data.Thread
		inf.meta.RespTCPSeq = data.TCPSeq

		inf.dur[1] = data.TS
		inf.ktime[2] = data.TSTail
	}

	switch dec.direction { //nolint:exhaustive
	case comm.DIn:
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			inf.reqBytes += data.FnCallSize
		case comm.NICDEgress:
			inf.respBytes += data.FnCallSize
		}
	case comm.DOut:
		switch txRx { //nolint:exhaustive
		case comm.NICDIngress:
			inf.respBytes += data.FnCallSize
		case comm.NICDEgress:
			inf.reqBytes += data.FnCallSize
		}
	}

	switch dec.reqResp {
	case 1:
		inf.ktime[1] = data.TSTail
	case 2:
		inf.ktime[3] = data.TSTail
	}
}

func (dec *pgsqlDecPipe) Proto() L7Protocol {
	return ProtoPgsql
}

func (dec *pgsqlDecPipe) Export(force bool) []*ProtoData {
	if force {
		dec.infCache = append(dec.infCache, dec.inf)
		dec.inf = nil
	}
	var result []*ProtoData

	for _, inf := range dec.infCache {
		if inf == nil || inf.resource == "" {
			continue
		}
		kvs := make(point.KVs, 0, 20)

		switch dec.direction { //nolint:exhaustive
		case comm.DIn:
			kvs = kvs.Set(comm.FieldBytesRead, int64(inf.reqBytes))
			kvs = kvs.Set(comm.FieldBytesWritten, int64(inf.respBytes))
		default:
			kvs = kvs.Set(comm.FieldBytesRead, int64(inf.respBytes))
			kvs = kvs.Set(comm.FieldBytesWritten, int64(inf.reqBytes))
		}
		kvs = kvs.Set(comm.FieldResource, inf.resource)
		kvs = kvs.Set(comm.FieldResourceType, inf.resourceType)
		kvs = kvs.Set(comm.FieldPgsqlComment, inf.comment)
		kvs = kvs.Set(comm.FieldPgsqlAffectedRow, inf.affectedRow)
		kvs = kvs.Set(comm.FieldPgsqlStatusCode, inf.statusCode)
		kvs = kvs.Set(comm.FieldStatus, toStatus(inf.status))
		kvs = kvs.Set(comm.FieldStatusI, status2String(inf.status))
		kvs = kvs.Set(comm.FieldPgsqlErrMsg, inf.errMsg)
		kvs = kvs.Set(comm.FieldPgsqlMsg, inf.msg)
		kvs = kvs.Set(comm.FieldPgsqlReqType, string(inf.reqType))
		kvs = kvs.Set(comm.FieldPgsqlRespType, string(inf.respType))
		kvs = kvs.Set(comm.FieldOperation, dec.Proto().String())

		dur := int64(inf.ktime[3] - inf.ktime[0])
		if dur <= 0 {
			dur = int64(inf.ktime[1] - inf.dur[0])
		}
		cost := int64(inf.ktime[2] - inf.ktime[1])
		if cost <= 0 {
			cost = int64(inf.ktime[1] - inf.dur[0])
		}
		result = append(result, &ProtoData{
			Meta:      inf.meta,
			Time:      inf.ts,
			KVs:       kvs,
			Cost:      cost,
			Duration:  dur,
			Direction: dec.direction,
			L7Proto:   ProtoPgsql,
			KTime:     inf.ktime[0],
		})
	}

	dec.infCache = dec.infCache[:0]
	return result
}

func (dec *pgsqlDecPipe) ConnClose() {
	dec.connClosed = true
}

func newPgsqlDecPipe(L7Protocol) ProtoDecPipe {
	return &pgsqlDecPipe{}
}

func (p *pgsqlInfo) isPgsql(sql []byte) bool {
	_, command, _ := trimCommentGetFirst(sql, 12)
	return p.isValidSQL(command)
}

func (p *pgsqlInfo) isValidSQL(word []byte) bool {
	if _, ok := sqlCommand[strings.ToUpper(string(word))]; ok {
		return true
	}
	if _, ok := pgsqlStart[strings.ToUpper(string(word))]; ok {
		return true
	}
	return false
}

func (dec *pgsqlDecPipe) determineDirection(port uint32, inf *pgsqlInfo) (direction Direction) {
	if port == dec.firstDstPort {
		direction = clientToServer
		if dec.reqResp == 2 && dec.inf.readyToExport {
			dec.infCache = append(dec.infCache, inf)
			inf = &pgsqlInfo{}
			dec.inf = inf
		}
	} else {
		direction = serverToClient
	}
	return direction
}

// readPgsqlBlock get Type and Real Length(except Type and Length).
func readPgsqlBlock(payload []byte) (byte, int, error) {
	if len(payload) < PgsqlMinLen {
		return 0, 0, fmt.Errorf("block size should gt %d", PgsqlMinLen)
	}

	tag := payload[0]
	length := int(binary.BigEndian.Uint32(payload[1:]))
	if length < PgsqlBlockMinLen || length+1 > len(payload) {
		return 0, 0, errors.New("block length error")
	}
	return tag, length - PgsqlTypeLenOffest, nil
}

const (
	PgsqlMinLen        = 5
	PgsqlBlockMinLen   = 4
	PgsqlTypeOffset    = 1
	PgsqlTypeLenOffest = 4
)

const (
	StatusOK = iota
	ClientError
	ServerError
	NotExistError
)

var pgsqlStart = map[string]struct{}{
	"WITH":         {},
	"EXECUTE":      {},
	"DECLARE":      {},
	"MATERIALIZED": {},
	"ABORT":        {},
	"ANALYZE":      {},
	"LOAD":         {},
	"LOCK":         {},
	"CHECKPOINT":   {},
	"REFRESH":      {},
	"REINDEX":      {},
	"RESET":        {},
	"START":        {},
	"TABLE":        {},
	"CLUSTER":      {},
	"TRUNCATE":     {},
}

var ErrUnknownProto = errors.New("unknown protocol")

func status2String(status int) string {
	switch status {
	case StatusOK:
		return "OK"
	case ClientError:
		return "ClientError"
	case ServerError:
		return "ServerError"
	case NotExistError:
		return "NotExistError"
	default:
		return ""
	}
}

func toStatus(status int) string {
	switch status {
	case StatusOK:
		return "ok"
	case ClientError, ServerError, NotExistError:
		return "error"
	default:
		return ""
	}
}

//nolint:funlen
func code2Msg(code string) (string, int) {
	switch code {
	// Client errors
	case "03000":
		return "sql_statement_not_yet_complete", ClientError
	case "0A000":
		return "feature_not_supported", ClientError
	case "0B000":
		return "invalid_transaction_initiation", ClientError
	case "0F000":
		return "locator_exception", ClientError
	case "0F001":
		return "invalid_locator_specification", ClientError
	case "0L000":
		return "invalid_grantor", ClientError
	case "0LP01":
		return "invalid_grant_operation", ClientError
	case "0P000":
		return "invalid_role_specification", ClientError
	case "20000":
		return "case_not_found", ClientError
	case "22000":
		return "data_exception", ClientError
	case "2202E":
		return "array_subscript_error", ClientError
	case "22021":
		return "character_not_in_repertoire", ClientError
	case "22008":
		return "datetime_field_overflow", ClientError
	case "22012":
		return "division_by_zero", ClientError
	case "22005":
		return "error_in_assignment", ClientError
	case "2200B":
		return "escape_character_conflict", ClientError
	case "22022":
		return "indicator_overflow", ClientError
	case "22015":
		return "interval_field_overflow", ClientError
	case "2201E":
		return "invalid_argument_for_logarithm", ClientError
	case "22014":
		return "invalid_argument_for_ntile_function", ClientError
	case "22016":
		return "invalid_argument_for_nth_value_function", ClientError
	case "2201F":
		return "invalid_argument_for_power_function", ClientError
	case "2201G":
		return "invalid_argument_for_width_bucket_function", ClientError
	case "22018":
		return "invalid_character_value_for_cast", ClientError
	case "22007":
		return "invalid_datetime_format", ClientError
	case "22019":
		return "invalid_escape_character", ClientError
	case "2200D":
		return "invalid_escape_octet", ClientError
	case "22025":
		return "invalid_escape_sequence", ClientError
	case "22P06":
		return "nonstandard_use_of_escape_character", ClientError
	case "22010":
		return "invalid_indicator_parameter_value", ClientError
	case "22023":
		return "invalid_parameter_value", ClientError
	case "22013":
		return "invalid_preceding_or_following_size", ClientError
	case "2201B":
		return "invalid_regular_expression", ClientError
	case "2201W":
		return "invalid_row_count_in_limit_clause", ClientError
	case "2201X":
		return "invalid_row_count_in_result_offset_clause", ClientError
	case "2202H":
		return "invalid_tablesample_argument", ClientError
	case "2202G":
		return "invalid_tablesample_repeat", ClientError
	case "22009":
		return "invalid_time_zone_displacement_value", ClientError
	case "2200C":
		return "invalid_use_of_escape_character", ClientError
	case "2200G":
		return "most_specific_type_mismatch", ClientError
	case "22004":
		return "null_value_not_allowed", ClientError
	case "22002":
		return "null_value_no_indicator_parameter", ClientError
	case "22003":
		return "numeric_value_out_of_range", ClientError
	case "2200H":
		return "sequence_generator_limit_exceeded", ClientError
	case "22026":
		return "string_data_length_mismatch", ClientError
	case "22001":
		return "string_data_right_truncation", ClientError
	case "22011":
		return "substring_error", ClientError
	case "22027":
		return "trim_error", ClientError
	case "22024":
		return "unterminated_c_string", ClientError
	case "2200F":
		return "zero_length_character_string", ClientError
	case "22P01":
		return "floating_point_exception", ClientError
	case "22P02":
		return "invalid_text_representation", ClientError
	case "22P03":
		return "invalid_binary_representation", ClientError
	case "22P04":
		return "bad_copy_file_format", ClientError
	case "22P05":
		return "untranslatable_character", ClientError
	case "2200L":
		return "not_an_xml_document", ClientError
	case "2200M":
		return "invalid_xml_document", ClientError
	case "2200N":
		return "invalid_xml_content", ClientError
	case "2200S":
		return "invalid_xml_comment", ClientError
	case "2200T":
		return "invalid_xml_processing_instruction", ClientError
	case "22030":
		return "duplicate_json_object_key_value", ClientError
	case "22031":
		return "invalid_argument_for_sql_json_datetime_function", ClientError
	case "22032":
		return "invalid_json_text", ClientError
	case "22033":
		return "invalid_sql_json_subscript", ClientError
	case "22034":
		return "more_than_one_sql_json_item", ClientError
	case "22035":
		return "no_sql_json_item", ClientError
	case "22036":
		return "non_numeric_sql_json_item", ClientError
	case "22037":
		return "non_unique_keys_in_a_json_object", ClientError
	case "22038":
		return "singleton_sql_json_item_required", ClientError
	case "22039":
		return "sql_json_array_not_found", ClientError
	case "2203A":
		return "sql_json_member_not_found", ClientError
	case "2203B":
		return "sql_json_number_not_found", ClientError
	case "2203C":
		return "sql_json_object_not_found", ClientError
	case "2203D":
		return "too_many_json_array_elements", ClientError
	case "2203E":
		return "too_many_json_object_members", ClientError
	case "2203F":
		return "sql_json_scalar_required", ClientError
	case "2203G":
		return "sql_json_item_cannot_be_cast_to_target_type", ClientError
	case "23000":
		return "integrity_constraint_violation", ClientError
	case "23001":
		return "restrict_violation", ClientError
	case "23502":
		return "not_null_violation", ClientError
	case "23503":
		return "foreign_key_violation", ClientError
	case "23505":
		return "unique_violation", ClientError
	case "23514":
		return "check_violation", ClientError
	case "23P01":
		return "exclusion_violation", ClientError
	case "26000":
		return "invalid_sql_statement_name", ClientError
	case "2F000":
		return "sql_routine_exception", ClientError
	case "2F005":
		return "function_executed_no_return_statement", ClientError
	case "2F002":
		return "modifying_sql_data_not_permitted", ClientError
	case "2F003":
		return "prohibited_sql_statement_attempted", ClientError
	case "2F004":
		return "reading_sql_data_not_permitted", ClientError
	case "34000":
		return "invalid_cursor_name", ClientError
	case "3D000":
		return "invalid_catalog_name", ClientError
	case "3F000":
		return "invalid_schema_name", ClientError
	case "40001":
		return "serialization_failure", ClientError
	case "40003":
		return "statement_completion_unknown", ClientError
	case "40P01":
		return "deadlock_detected", ClientError
	case "42000":
		return "syntax_error_or_access_rule_violation", ClientError
	case "42601":
		return "syntax_error", ClientError
	case "42501":
		return "insufficient_privilege", ClientError
	case "42846":
		return "cannot_coerce", ClientError
	case "42803":
		return "grouping_error", ClientError
	case "42P20":
		return "windowing_error", ClientError
	case "42P19":
		return "invalid_recursion", ClientError
	case "42830":
		return "invalid_foreign_key", ClientError
	case "42602":
		return "invalid_name", ClientError
	case "42622":
		return "name_too_long", ClientError
	case "42939":
		return "reserved_name", ClientError
	case "42804":
		return "datatype_mismatch", ClientError
	case "42P18":
		return "indeterminate_datatype", ClientError
	case "42P21":
		return "collation_mismatch", ClientError
	case "42P22":
		return "indeterminate_collation", ClientError
	case "42809":
		return "wrong_object_type", ClientError
	case "428C9":
		return "generated_always", ClientError
	case "42703":
		return "undefined_column", ClientError
	case "42883":
		return "undefined_function", ClientError
	case "42P01":
		return "undefined_table", ClientError
	case "42P02":
		return "undefined_parameter", ClientError
	case "42704":
		return "undefined_object", ClientError
	case "42701":
		return "duplicate_column", ClientError
	case "42P03":
		return "duplicate_cursor", ClientError
	case "42P04":
		return "duplicate_database", ClientError
	case "42723":
		return "duplicate_function", ClientError
	case "42P05":
		return "duplicate_prepared_statement", ClientError
	case "42P06":
		return "duplicate_schema", ClientError
	case "42P07":
		return "duplicate_table", ClientError
	case "42712":
		return "duplicate_alias", ClientError
	case "42710":
		return "duplicate_object", ClientError
	case "42702":
		return "ambiguous_column", ClientError
	case "42725":
		return "ambiguous_function", ClientError
	case "42P08":
		return "ambiguous_parameter", ClientError
	case "42P09":
		return "ambiguous_alias", ClientError
	case "42P10":
		return "invalid_column_reference", ClientError
	case "42611":
		return "invalid_column_definition", ClientError
	case "42P11":
		return "invalid_cursor_definition", ClientError
	case "42P12":
		return "invalid_database_definition", ClientError
	case "42P13":
		return "invalid_function_definition", ClientError
	case "42P14":
		return "invalid_prepared_statement_definition", ClientError
	case "42P15":
		return "invalid_schema_definition", ClientError
	case "42P16":
		return "invalid_table_definition", ClientError
	case "42P17":
		return "invalid_object_definition", ClientError

	// Server errors
	case "0Z000":
		return "diagnostics_exception", ServerError
	case "0Z002":
		return "stacked_diagnostics_accessed_without_active_handler", ServerError
	case "08000":
		return "connection_exception", ServerError
	case "08003":
		return "connection_does_not_exist", ServerError
	case "08006":
		return "connection_failure", ServerError
	case "08001":
		return "sqlclient_unable_to_establish_sqlconnection", ServerError
	case "08004":
		return "sqlserver_rejected_establishment_of_sqlconnection", ServerError
	case "08007":
		return "transaction_resolution_unknown", ServerError
	case "08P01":
		return "protocol_violation", ServerError
	case "09000":
		return "triggered_action_exception", ServerError
	case "21000":
		return "cardinality_violation", ServerError
	case "24000":
		return "invalid_cursor_state", ServerError
	case "25000":
		return "invalid_transaction_state", ServerError
	case "25001":
		return "active_sql_transaction", ServerError
	case "25002":
		return "branch_transaction_already_active", ServerError
	case "25008":
		return "held_cursor_requires_same_isolation_level", ServerError
	case "25003":
		return "inappropriate_access_mode_for_branch_transaction", ServerError
	case "25004":
		return "inappropriate_isolation_level_for_branch_transaction", ServerError
	case "25005":
		return "no_active_sql_transaction_for_branch_transaction", ServerError
	case "25006":
		return "read_only_sql_transaction", ServerError
	case "25007":
		return "schema_and_data_statement_mixing_not_supported", ServerError
	case "25P01":
		return "no_active_sql_transaction", ServerError
	case "25P02":
		return "in_failed_sql_transaction", ServerError
	case "25P03":
		return "idle_in_transaction_session_timeout", ServerError
	case "27000":
		return "triggered_data_change_violation", ServerError
	case "28000":
		return "invalid_authorization_specification", ServerError
	case "28P01":
		return "invalid_password", ServerError
	case "2B000":
		return "dependent_privilege_descriptors_still_exist", ServerError
	case "2BP01":
		return "dependent_objects_still_exist", ServerError
	case "2D000":
		return "invalid_transaction_termination", ServerError
	case "38000":
		return "external_routine_exception", ServerError
	case "38001":
		return "containing_sql_not_permitted", ServerError
	case "38002":
		return "modifying_sql_data_not_permitted", ServerError
	case "38003":
		return "prohibited_sql_statement_attempted", ServerError
	case "38004":
		return "reading_sql_data_not_permitted", ServerError
	case "39000":
		return "external_routine_invocation_exception", ServerError
	case "39001":
		return "invalid_sqlstate_returned", ServerError
	case "39004":
		return "null_value_not_allowed", ServerError
	case "39P01":
		return "trigger_protocol_violated", ServerError
	case "39P02":
		return "srf_protocol_violated", ServerError
	case "39P03":
		return "event_trigger_protocol_violated", ServerError
	case "3B000":
		return "savepoint_exception", ServerError
	case "3B001":
		return "invalid_savepoint_specification", ServerError
	case "40000":
		return "transaction_rollback", ServerError
	case "40002":
		return "transaction_integrity_constraint_violation", ServerError
	case "44000":
		return "with_check_option_violation", ServerError
	case "53000":
		return "insufficient_resources", ServerError
	case "53100":
		return "disk_full", ServerError
	case "53200":
		return "out_of_memory", ServerError
	case "53300":
		return "too_many_connections", ServerError
	case "53400":
		return "configuration_limit_exceeded", ServerError
	case "54000":
		return "program_limit_exceeded", ServerError
	case "54001":
		return "statement_too_complex", ServerError
	case "54011":
		return "too_many_columns", ServerError
	case "54023":
		return "too_many_arguments", ServerError
	case "55000":
		return "object_not_in_prerequisite_state", ServerError
	case "55006":
		return "object_in_use", ServerError
	case "55P02":
		return "cant_change_runtime_param", ServerError
	case "55P03":
		return "lock_not_available", ServerError
	case "55P04":
		return "unsafe_new_enum_value_usage", ServerError
	case "57000":
		return "operator_intervention", ServerError
	case "57014":
		return "query_canceled", ServerError
	case "57P01":
		return "admin_shutdown", ServerError
	case "57P02":
		return "crash_shutdown", ServerError
	case "57P03":
		return "cannot_connect_now", ServerError
	case "57P04":
		return "database_dropped", ServerError
	case "57P05":
		return "idle_session_timeout", ServerError
	case "58000":
		return "system_error", ServerError
	case "58030":
		return "io_error", ServerError
	case "58P01":
		return "undefined_file", ServerError
	case "58P02":
		return "duplicate_file", ServerError
	case "72000":
		return "snapshot_too_old", ServerError
	case "F0000":
		return "config_file_error", ServerError
	case "F0001":
		return "lock_file_exists", ServerError
	case "HV000":
		return "fdw_error", ServerError
	case "HV005":
		return "fdw_column_name_not_found", ServerError
	case "HV002":
		return "fdw_dynamic_parameter_value_needed", ServerError
	case "HV010":
		return "fdw_function_sequence_error", ServerError
	case "HV021":
		return "fdw_inconsistent_descriptor_information", ServerError
	case "HV024":
		return "fdw_invalid_attribute_value", ServerError
	case "HV007":
		return "fdw_invalid_column_name", ServerError
	case "HV008":
		return "fdw_invalid_column_number", ServerError
	case "HV004":
		return "fdw_invalid_data_type", ServerError
	case "HV006":
		return "fdw_invalid_data_type_descriptors", ServerError
	case "HV091":
		return "fdw_invalid_descriptor_field_identifier", ServerError
	case "HV00B":
		return "fdw_invalid_handle", ServerError
	case "HV00C":
		return "fdw_invalid_option_index", ServerError
	case "HV00D":
		return "fdw_invalid_option_name", ServerError
	case "HV090":
		return "fdw_invalid_string_length_or_buffer_length", ServerError
	case "HV00A":
		return "fdw_invalid_string_format", ServerError
	case "HV009":
		return "fdw_invalid_use_of_null_pointer", ServerError
	case "HV014":
		return "fdw_too_many_handles", ServerError
	case "HV001":
		return "fdw_out_of_memory", ServerError
	case "HV00P":
		return "fdw_no_schemas", ServerError
	case "HV00J":
		return "fdw_option_name_not_found", ServerError
	case "HV00K":
		return "fdw_reply_handle", ServerError
	case "HV00Q":
		return "fdw_schema_not_found", ServerError
	case "HV00R":
		return "fdw_table_not_found", ServerError
	case "HV00L":
		return "fdw_unable_to_create_execution", ServerError
	case "HV00M":
		return "fdw_unable_to_create_reply", ServerError
	case "HV00N":
		return "fdw_unable_to_establish_connection", ServerError
	case "P0000":
		return "plpgsql_error", ServerError
	case "P0001":
		return "raise_exception", ServerError
	case "P0002":
		return "no_data_found", ServerError
	case "P0003":
		return "too_many_rows", ServerError
	case "P0004":
		return "assert_failure", ServerError
	case "XX000":
		return "internal_error", ServerError
	case "XX001":
		return "data_corrupted", ServerError
	case "XX002":
		return "index_corrupted", ServerError
	default:
		return "", NotExistError
	}
}
