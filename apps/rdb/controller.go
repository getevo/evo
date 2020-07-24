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
	if v, ok := dbo.Queries[name]; ok {
		return v
	}
	log.Error("query %s not found", name)
	return nil
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
		if request, ok := params[0].(*evo.Request); ok && q.Parser != nil {
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

func (q *Query) ToMap(params ...interface{}) ([]map[string]interface{}, error) {
	var out []map[string]interface{}
	if len(params) == 1 {

		if request, ok := params[0].(*evo.Request); ok {
			parameters, err := q.Parser.Parse(request)
			if err != nil {
				return out, err
			}
			var parametersInterface []interface{}
			for _, item := range parameters {
				parametersInterface = append(parametersInterface, item)
			}
			return q.ToMap(parametersInterface...)
		}
	}
	var err error
	var rows *sql.Rows
	/*	if q.DBO.Debug {
		//log.Infof(strings.Replace(q.QueryString, "?", "%s", len(params)), params...)
	}*/

	if rows, err = q.DBO.Conn.Query(q.QueryString, params...); err != nil {
		return out, err
	}

	cols, _ := rows.Columns()

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return out, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		out = append(out, m)

	}
	return out, nil

}
