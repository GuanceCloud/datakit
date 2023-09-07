package go_ibm_db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
        "time"
        "sync"
)

//DBP struct type contains the timeout, dbinstance and connection string
type DBP struct {
	sql.DB
	con string
	n   time.Duration
}

//Pool struct contais the about the pool like size, used and available connections
type Pool struct {
	availablePool map[string][]*DBP
	usedPool      map[string][]*DBP
	poolSize      int
        mu            sync.Mutex
}

var b *Pool
var connMaxLifetime, poolSize int
const defaultMaxIdleConns = 2
const defaultConnMaxLifetime = 60

//Pconnect will return the pool instance
func Pconnect(poolSize string) *Pool {
	var size int
	count := len(poolSize)
	if count > 0 {
		opt := strings.Split(poolSize, "=")
		if opt[0] == "PoolSize" {
			size, _ = strconv.Atoi(opt[1])
			if size <= 0 {
				size = defaultMaxIdleConns
			}
		} else {
			fmt.Println("Not a valid parameter")
		}
	} else {
		size = defaultMaxIdleConns
	}
	p := &Pool{
		availablePool: make(map[string][]*DBP),
		usedPool:      make(map[string][]*DBP),
		poolSize:      size,
	}
	b = p

	return p
}

//Psize sets the size of the pool idf value is passed
var pSize int

//Open will check for the connection in the pool
//If not opens a new connection and stores in the pool
func (p *Pool) Open(connStr string, options ...string) *DBP {
	var Time time.Duration
	count := len(options)
	if count > 0 {
		for i := 0; i < count; i++ {
			opt := strings.Split(options[i], "=")
			if opt[0] == "SetConnMaxLifetime" {
				connMaxLifetime, _ = strconv.Atoi(opt[1])
				if connMaxLifetime <= 0 {
					connMaxLifetime = defaultConnMaxLifetime
				}
				Time = time.Duration(connMaxLifetime) * time.Second
			} else {
				fmt.Println("Not a valid parameter")
			}
		}
	} else {
		Time = time.Duration(defaultConnMaxLifetime) * time.Second
	}
	if pSize < p.poolSize {
		pSize = pSize + 1
		if val, ok := p.availablePool[connStr]; ok {
			if len(val) > 1 {
				p.mu.Lock()
				dbpo := val[0]
				copy(val[0:], val[1:])
				val[len(val)-1] = nil
				val = val[:len(val)-1]
				p.availablePool[connStr] = val
				p.usedPool[connStr] = append(p.usedPool[connStr], dbpo)
				dbpo.SetConnMaxLifetime(Time)
				p.mu.Unlock()
				return dbpo
			} else {
				p.mu.Lock()
				dbpo := val[0]
				p.usedPool[connStr] = append(p.usedPool[connStr], dbpo)
				delete(p.availablePool, connStr)
				dbpo.SetConnMaxLifetime(Time)
				p.mu.Unlock()
				return dbpo
			}
		} else {
			db, err := sql.Open("go_ibm_db", connStr)
			if err != nil {
				return nil
			}
			dbi := &DBP{
				DB:  *db,
				con: connStr,
				n:   Time,
			}
			p.mu.Lock()
			p.usedPool[connStr] = append(p.usedPool[connStr], dbi)
			dbi.SetConnMaxLifetime(Time)
			p.mu.Unlock()
			return dbi
		}
        } else {
		pSize = pSize + 1

		for i := 0; i < connMaxLifetime; i++ {
			if len(p.availablePool) <= 0 {
				time.Sleep(3 * time.Second)
				i = i + 2
			} else {
				if val, ok := p.availablePool[connStr]; ok {
					if len(val) > 1 {
						p.mu.Lock()
						dbpo := val[0]
						copy(val[0:], val[1:])
						val[len(val)-1] = nil
						val = val[:len(val)-1]
						p.availablePool[connStr] = val
						p.usedPool[connStr] = append(p.usedPool[connStr], dbpo)
						dbpo.SetConnMaxLifetime(Time)
						p.mu.Unlock()
						return dbpo
					} else {
						dbpo :=  val[0]
						return dbpo
					}
				}
			}
		}
		fmt.Println("Connection timeout")
		return nil
	}
	return nil
}

func (p *Pool) Init(numConn int, connStr string) bool{
	var Time time.Duration

	if  connMaxLifetime  <= 0 {
		Time = time.Duration(defaultConnMaxLifetime) * time.Second
	} else {
		Time = time.Duration(connMaxLifetime) * time.Second
	}

	for i := 0; i < numConn; i++ {
		db, err := sql.Open("go_ibm_db", connStr)
		if err != nil {
			return false
		}
		dbi := &DBP{
			DB:  *db,
			con: connStr,
			n:   Time,
		}
		p.mu.Lock()
		p.availablePool[connStr] = append(p.availablePool[connStr], dbi)
		dbi.SetConnMaxLifetime(Time)
		p.mu.Unlock()
	}
	return true
}

//Close will make the connection available for the next release
func (d *DBP) Close() {
	pSize = pSize - 1
	var pos int
	i := -1
        b.mu.Lock()
	if valc, okc := b.usedPool[d.con]; okc {
		if len(valc) > 1 {
			for _, b := range valc {
				i = i + 1
				if b == d {
					pos = i
				}
			}
			dbpc := valc[pos]
			copy(valc[pos:], valc[pos+1:])
			valc[len(valc)-1] = nil
			valc = valc[:len(valc)-1]
			b.usedPool[d.con] = valc
			b.availablePool[d.con] = append(b.availablePool[d.con], dbpc)
		} else {
			dbpc := valc[0]
			b.availablePool[d.con] = append(b.availablePool[d.con], dbpc)
			delete(b.usedPool, d.con)
		}
		go d.Timeout()
	} else {
		d.DB.Close()
	}
         b.mu.Unlock()
}

//Timeout for closing the connection in pool
func (d *DBP) Timeout() {
	var pos int
	i := -1
	select {
	case <-time.After(d.n):
                b.mu.Lock()
		if valt, okt := b.availablePool[d.con]; okt {
			if len(valt) > 1 {
				for _, b := range valt {
					i = i + 1
					if b == d {
						pos = i
					}
				}
				dbpt := valt[pos]
				copy(valt[pos:], valt[pos+1:])
				valt[len(valt)-1] = nil
				valt = valt[:len(valt)-1]
				b.availablePool[d.con] = valt
				dbpt.DB.Close()
			} else {
				dbpt := valt[0]
				dbpt.DB.Close()
				delete(b.availablePool, d.con)
			}
		}
                b.mu.Unlock()
	}
}

//Release will close all the connections in the pool
func (p *Pool) Release() {
	if p.availablePool != nil {
		for _, vala := range p.availablePool {
			for _, dbpr := range vala {
				dbpr.DB.Close()
			}
		}
		p.availablePool = nil
	}
	if p.usedPool != nil {
		for _, valu := range p.usedPool {
			for _, dbpr := range valu {
				dbpr.DB.Close()
			}
		}
		p.usedPool = nil
	}
}

//Set the connMaxLifetime
func (p *Pool) SetConnMaxLifetime(num int) {
	connMaxLifetime = num
}

// Display will print the  values in the map
func (p *Pool) Display() {
	fmt.Println(p.availablePool)
	fmt.Println(p.usedPool)
	fmt.Println(p.poolSize)
}
