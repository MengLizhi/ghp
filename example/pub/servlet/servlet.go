package main

import (
  "ghp"
  "fmt"
  "./foo"
  "./bar"
)

var ctx ghp.ServletContext

func StartServlet(c ghp.ServletContext) {
  ctx = c
  println("starting servlet instance " + ctx.Version())
}

func StopServlet(c ghp.ServletContext) {
  println("stopping servlet instance " + ctx.Version())
}

func ServeHTTP(r *ghp.Request, w ghp.Response) {
  query := r.URL.Query()

  fmt.Fprintf(w, "Request URL query: %+v\n", query)

  w.WriteString("Hello from a servlet.\n")
  w.WriteString("foo.Foo() =>                " + foo.Foo() + "\n")
  w.WriteString("bar.Bar() =>                " + bar.Bar() + "\n")
  w.WriteString("ServletContext.Name() =>    " + ctx.Name() + "\n")
  w.WriteString("ServletContext.Version() => " + ctx.Version() + "\n")
}
