package rdb

import (
	"database/sql"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackskj/carta"
	"strings"
	"sync"
)

type Controller struct{}

type DBO struct {
	Conn    *sql.DB
	Queries map[string]*Query
	Debug   bool
	mu      sync.Mutex
}

type Query struct {
	QueryString string
	DBO         *DBO
	Parser      *Parser
}

func CreateConnections(config map[string]string) error {
	var err error
	for name, dsn := range config {
		_, err = CreateConnection(name, "mysql", dsn)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateConnection(name, driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return db, err
	}
	return db, PushDB(name, db)
}

func PushDB(name string, db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}
	connections.Set(name, &DBO{
		db, map[string]*Query{}, false, sync.Mutex{},
	})
	return nil
}

func GetDBO(name string) *DBO {
	if connections.Has(name) {
		return connections.Get(name).(*DBO)
	}
	return nil
}

func (dbo *DBO) GetQuery(name string) *Query {
	dbo.mu.Lock()
	defer dbo.mu.Unlock()
	return dbo.Queries[name]
}

func (dbo *DBO) CreateQuery(name, query string) *Query {
	dbo.mu.Lock()
	defer dbo.mu.Unlock()
	q := &Query{query, dbo, nil}
	dbo.Queries[name] = q
	return q
}

func (q *Query) SetParser(parser *Parser) {
	q.Parser = parser
}

func (q *Query) All(out interface{}, params ...interface{}) error {

	if len(params) == 1 {
		if request, ok := params[0].(*evo.Request); ok {
			parameters, err := q.Parser.Parse(request)
			if err != nil {
				return err
			}
			var parametersInterface []interface{}
			for _, item := range parameters {
				parametersInterface = append(parametersInterface, item)
			}
			return q.All(out, parametersInterface...)
		}
	}
	var err error
	var rows *sql.Rows
	if q.DBO.Debug {
		log.Infof(strings.Replace(q.QueryString, "?", "%s", len(params)), params...)
	}
	if rows, err = q.DBO.Conn.Query(q.QueryString, params...); err != nil {
		return err
	}
	return carta.Map(rows, out)
}
