package gist
import (
"fmt"
"io"
)
//line index.ego:1
 func (_ *tmpl) Gists(w io.Writer, gists []*Gist) error  {
//line index.ego:2
if _, err := fmt.Fprintf(w, "\n\n"); err != nil { return err }
//line index.ego:3
if _, err := fmt.Fprintf(w, "<!DOCTYPE html>\n"); err != nil { return err }
//line index.ego:4
if _, err := fmt.Fprintf(w, "<html lang=\"en\">\n  "); err != nil { return err }
//line index.ego:5
if _, err := fmt.Fprintf(w, "<body>\n    "); err != nil { return err }
//line index.ego:6
if _, err := fmt.Fprintf(w, "<h1>Gist Exposed!"); err != nil { return err }
//line index.ego:6
if _, err := fmt.Fprintf(w, "</h1>\n\n    "); err != nil { return err }
//line index.ego:8
if _, err := fmt.Fprintf(w, "<h2>Recent Gists"); err != nil { return err }
//line index.ego:8
if _, err := fmt.Fprintf(w, "</h2>\n\n    "); err != nil { return err }
//line index.ego:10
if _, err := fmt.Fprintf(w, "<ol>\n      "); err != nil { return err }
//line index.ego:11
 for _, g := range gists { 
//line index.ego:12
if _, err := fmt.Fprintf(w, "\n        "); err != nil { return err }
//line index.ego:12
if _, err := fmt.Fprintf(w, "<li>\n          "); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "<a href=\"/"); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "%v", g.Owner); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "/"); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "%v", g.ID); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "\">\n            "); err != nil { return err }
//line index.ego:14
 if g.Description != "" { 
//line index.ego:15
if _, err := fmt.Fprintf(w, "\n              "); err != nil { return err }
//line index.ego:15
if _, err := fmt.Fprintf(w, "%v",  g.Description ); err != nil { return err }
//line index.ego:16
if _, err := fmt.Fprintf(w, "\n            "); err != nil { return err }
//line index.ego:16
 } else { 
//line index.ego:17
if _, err := fmt.Fprintf(w, "\n              "); err != nil { return err }
//line index.ego:17
if _, err := fmt.Fprintf(w, "<em>Untitled"); err != nil { return err }
//line index.ego:17
if _, err := fmt.Fprintf(w, "</em>\n            "); err != nil { return err }
//line index.ego:18
 } 
//line index.ego:19
if _, err := fmt.Fprintf(w, "\n          "); err != nil { return err }
//line index.ego:19
if _, err := fmt.Fprintf(w, "</a>\n        "); err != nil { return err }
//line index.ego:20
if _, err := fmt.Fprintf(w, "</li>\n      "); err != nil { return err }
//line index.ego:21
 } 
//line index.ego:22
if _, err := fmt.Fprintf(w, "\n    "); err != nil { return err }
//line index.ego:22
if _, err := fmt.Fprintf(w, "</ol>\n\n  "); err != nil { return err }
//line index.ego:24
if _, err := fmt.Fprintf(w, "</body>\n"); err != nil { return err }
//line index.ego:25
if _, err := fmt.Fprintf(w, "</html>\n"); err != nil { return err }
return nil
}
