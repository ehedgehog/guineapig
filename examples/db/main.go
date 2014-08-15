package main

import "fmt"
import "database/sql"
import "strings"
import "net/http"
import "io/ioutil"

// import "github.com/ziutek/mymysql/mysql"
// import _ "github.com/ziutek/mymysql/native"

import _ "github.com/go-sql-driver/mysql"

func lookupPage(db *sql.DB, out http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(4096)
	file, header, err := req.FormFile("upload")

	fmt.Println("header:", header.Filename, "&", header.Header)
	panicUnlessNil(err)

	contents, err := ioutil.ReadAll(file)
	panicUnlessNil(err)

	fmt.Println("file contents:", string(contents))

	fmt.Println(req.Form)

	lines := strings.Split(string(contents), "\n")

	out.Write([]byte(`<html><head></head><body>`))
	out.Write([]byte("\n<pre>\n"))

	for _, line := range lines {

		rows, err := db.Query("SELECT something FROM sometable WHERE CONCAT(foo, ' ', bar) = ?", line)
		panicUnlessNil(err)

		for rows.Next() {
			var something string
			rows.Scan(&postcode)
			fmt.Println("from", line, "produced", something)
			fmt.Fprintf(out, "%v => %v\n", line, something)
			break
		}
	}

	out.Write([]byte("\n"))
	out.Write([]byte(`</body></html>`))
	out.Write([]byte("\n"))

	// http.Redirect(out, req, "/", http.StatusFound)
}

var mainPageSource = `
	<html>
	<head>
	</head>
	<body>
	<h1>oopsypegs</h1>
	<form style="background: yellow" action="lookup" method="POST" enctype="multipart/form-data">
		<input style="border: 1px solid red" name="upload" type="file" size="40" />
		<input type="submit" name="henry" value="UPLOAD A" />
		<input type="submit" name="janny" value="UPLOAD B" />
	</form>
	</body>
	</html>
`

func mainPage(out http.ResponseWriter, req *http.Request) {
	out.Write([]byte(mainPageSource))
}

func main() {
	fmt.Println("here we go round the bush.")

	db, err := sql.Open("mysql", "USERNAME-AND-PASSWORD@/DATABASE?charset=utf8")
	panicUnlessNil(err)

	fmt.Println("we have opened the database.")

	http.Handle("/", http.HandlerFunc(mainPage))
	http.Handle("/lookup", http.HandlerFunc(func(out http.ResponseWriter, req *http.Request) { lookupPage(db, out, req) }))

	err = http.ListenAndServe(":8086", nil)
	panicUnlessNil(err)
}

func panicUnlessNil(err error) {
	if err != nil {
		panic(err)
	}
}
