package rdb

import (
	"database/sql"
	"github.com/getevo/evo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackskj/carta"
)

type Controller struct{}

type DBO struct {
	Conn *sql.DB
}

type Query struct {
	QueryString string
	DBO         *DBO
	Parser      *Parser
}

func CreateConnection(name, driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	return PushDB(name, db)
}

func PushDB(name string, db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return err
	}
	connections.Set(name, &DBO{
		db,
	})
	return nil
}

func GetDBO(name string) *DBO {
	if connections.Has(name) {
		return connections.Get(name).(*DBO)
	}
	return nil
}

func (dbo *DBO) Query(query string) *Query {
	return &Query{query, dbo, nil}
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
	if rows, err = q.DBO.Conn.Query(q.QueryString, params...); err != nil {
		return err
	}
	return carta.Map(rows, out)
}
