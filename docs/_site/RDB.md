RDB or Readonly Database is a fast-multi database access data reader and binder,  which let the developer to access multiple database connections at same time and query them to read data very fast then cast result to structured data.


** Note: RDB is not an ORM. RDB is used for api systems to make reusable raw queries and bind http request to query and cast result to struct without to much overhead **

Example:

1- Create connection and add to rdp pool for later use:

```go
rdb.Register()
err := rdb.CreateConnection("localhost", "mysql", "root:password@localhost/mydb?charset=utf8&parseTime=True")
if err != nil {
	log.Fatal(err)
}
```

2- Access db object:

```go
var db = rdb.GetDBO("localhost")
if db == nil {
    log.Fatal("Null db")
}
```

3- Define Reusable Query, http parser and struct:

for query and structure refer to 
[[jackskj/carta|https://github.com/jackskj/carta]]
 and for data validation refer to [[go-playground/validator|https://github.com/go-playground/validator]]
```go
query := db.Query(`
SELECT
       id          as  blog_id,
       title       as  blog_title,
       P.id        as  posts_id,         
       P.name      as  posts_name,
       A.id        as  author_id,      
       A.username  as  author_username
FROM blog
       left outer join author A    on  blog.author_id = A.id
       left outer join post P      on  blog.id = P.blog_id
WHERE blog.id_category = ?
`)

type Blog struct {
        Id     int    `db:"blog_id"`
        Title  string `db:"blog_title"`
        Posts  []Post
        Author Author
}
type Post struct {
        Id   int    `db:"posts_id"`
        Name string `db:"posts_name"`
}
type Author struct {
        Id       int    `db:"author_id"`
        Username string `db:"author_username"`
}

//optional: use http parser to auto acquire params from http call 
parser := &rdb.Parser{
	Params: []rdb.Param{
                // { parameter name , parameter source (URL,Get,Post,Header,Any) , validator based on https://github.com/go-playground/validator }
		{"id", rdb.URL, "numeric"},
	},
}

query.SetParser(parser)
```

4- Manipulating parsed params if needed:
```go
parser.Processor = func(params []string) []string {
	params[0] = strconv.Itoa( lib.ParseSafeInt(params[0]) + 1 ) // add 1 to first arg
	return params
}
```

5- Create API:
```go
//with parser
evo.Get("/api/blog/:id", func(request *evo.Request) {
	posts := []Blog{}
	err := query.All(&posts, request)
	if err != nil {
		log.Error(err)
	}
	request.WriteResponse(data)
})

// without http parser
evo.Get("/api/blog/:id", func(request *evo.Request) {
	posts := []Blog{}
	err := query.All(&posts, request.Param("id"))
	if err != nil {
		log.Error(err)
	}
	request.WriteResponse(data)
})
```

RDB will take argument, generate the sql and query to db then map the SQL rows while keeping track of those relationships. 

Results: 
```
rows:
blog_id | blog_title | posts_id | posts_name | author_id | author_username
1       | Foo        | 1        | Bar        | 1         | John
1       | Foo        | 2        | Baz        | 1         | John
2       | Egg        | 3        | Beacon     | 2         | Ed

blogs:
[{
	"blog_id": 1,
	"blog_title": "Foo",
	"author": {
		"author_id": 1,
		"author_username": "John"
	},
	"posts": [{
			"post_id": 1,
			"posts_name": "Bar"
		}, {
			"post_id": 2,
			"posts_name": "Baz"
		}]
}, {
	"blog_id": 2,
	"blog_title": "Egg",
	"author": {
		"author_id": 2,
		"author_username": "Ed"
	},
	"posts": [{
			"post_id": 3,
			"posts_name": "Beacon"
		}]
}]
```

**Column and Field Names**

RDB will match your SQL columns with corresponding fields. You can use a "db" tag to represent a specific column name.
Example:

```
type Blog struct {
	// When tag is not used, the snake case of the fiels is used
	BlogId int // expected column name : "blog_id"

	// When tag is specified, it takes priority
	Abc string `db:"blog_title"` // expected column name: "blog_title"

	// If you define multiple fiels with the same struct,
	// you can use a tag to identify a column prefix 
	// (with underscore concatination)

	// possible column names:  "writer_author_id", "author_id"
	Writer Author `db: "writer"`
        
	// possible column names: "rewiewer_author_id", "author_id",
	Reviewer Author `db: "reviewer"`
}

type Author struct {
	AuthorId int `db:"author_id"`
}
```

**Data Types and Relationships**

Any primative types, time.Time, protobuf Timestamp, and sql.NullX can be loaded.
These types are one-to-one mapped with your SQL columns

To define more complex SQL relationships use slices and structs as in example below:

```
type Blog struct {
	BlogId int  // Will map directly with "blog_id" column 

	// If your SQL data can be "null", use pointers or sql.NullX
	AuthorId  *int
	CreatedOn *timestamp.Timestamp // protobuf timestamp
	UpdatedOn *time.Time
	SonsorId  sql.NullInt64

	// To define has-one relationship, use nested structs 
	// or pointer to a struct
	Author *Author

	// To define has-many relationship, use slices
	// options include: *[]*Post, []*Post, *[]Post, []Post
	Posts []*Post 

	// If your has-many relationship corresponds to one column,
	// you can use a slice of a settable type
	TagIds     []int           `db:"tag_id"`
	CommentIds []sql.NullInt64 `db:"comment_id"`
}
```

**Important Note**

RDB removes any duplicate rows. This is a side effect of the data mapping as it is unclear which object to instantiate if the same data arrives more than once. If this is not a desired outcome, you should include a uniquely identifiable columns in your query and the corresponding fields in your structs.

To prevent relatively expensive reflect operations, RDB caches the structure of your struct using the column names of your query response as well as the type of your struct.


