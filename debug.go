package go_rpc

import (
	"fmt"
	"html/template"
	"net/http"
)

const debugText = `<!DOCTYPE html>
<html>
<head>
	<title>Go-RPC Services</title>
	<style>
		body {
			background-color: #f7f7f7;
			font-family: Arial, sans-serif;
			text-align: center;
		}

		h1 {
			margin-top: 50px;
			font-size: 48px;
			color: #333333;
			text-shadow: 2px 2px #dcdcdc;
		}

		h2 {
			margin-top: 40px;
			font-size: 24px;
			color: #666666;
			text-shadow: 1px 1px #dcdcdc;
		}

		table {
			margin: 0 auto;
			margin-top: 20px;
			border-collapse: collapse;
			box-shadow: 0px 0px 5px #aaaaaa;
			background-color: #e6f0ff;
		}

		th, td {
			padding: 10px;
			border: 1px solid #dddddd;
			font-family: monospace;
			font-size: 16px;
			color: #333333;
			background-color: #ffffff;
		}

		th {
			background-color: #f2f2f2;
			font-weight: bold;
			color: #555555;
			text-align: left;
			box-shadow: 0px 2px #aaaaaa;
		}

		td {
			box-shadow: 0px 2px #aaaaaa inset;
		}

		tr:nth-child(even) {
			background-color: #f9f9f9;
		}

		tr:hover {
			background-color: #f5f5f5;
		}

		#logo {
			position: absolute;
			top: 10px;
			right: 10px;
		}

		#logo img {
			height: 120px;
		}
	</style>
</head>
<body>
	<a href="https://golang.org/" id="logo"><img src="https://blog.golang.org/go-brand/Go-Logo/SVG/Go-Logo_Blue.svg" alt="Golang logo"></a>
	<h1>Go-RPC Services</h1>
	{{range .}}
	<hr>
	<h2>Service {{.Name}} ------------------ from: {{.ServerAdr}}</h2>
	<hr>
	<table>
		<tr>
			<th>Method</th>
			<th>Calls</th>
		</tr>
		{{range $name, $mtype := .Method}}
			<tr>
				<td>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
				<td>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
	</table>
	{{end}}
</body>
</html>
`

var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugHTTP struct {
	servers []*Server
}

type debugService struct {
	Name      string
	Method    map[string]*methodType
	ServerAdr string
}

func (d debugHTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Build a sorted version of the data.
	var services []debugService
	for i := 0; i < len(d.servers); i++ {
		d.servers[i].serviceMap.Range(func(namei, svci interface{}) bool {
			svc := svci.(*service)
			services = append(services, debugService{
				Name:      namei.(string),
				Method:    svc.method,
				ServerAdr: d.servers[i].bindAdr,
			})
			return true
		})
	}
	err := debug.Execute(w, services)
	if err != nil {
		_, _ = fmt.Fprintln(w, "rpc: error executing template:", err.Error())
	}
}
