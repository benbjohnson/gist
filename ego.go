package gist
import (
"fmt"
"io"
"time"
)
//line dashboard.ego:1
 func (t *tmpl) Dashboard(w io.Writer, hosted, recent []*Gist) error  {
//line dashboard.ego:2
if _, err := fmt.Fprintf(w, "\n\n"); err != nil { return err }
//line dashboard.ego:4
if _, err := fmt.Fprintf(w, "\n\n"); err != nil { return err }
//line dashboard.ego:5
if _, err := fmt.Fprintf(w, "<!DOCTYPE html>\n"); err != nil { return err }
//line dashboard.ego:6
if _, err := fmt.Fprintf(w, "<html lang=\"en\">\n  "); err != nil { return err }
//line dashboard.ego:7
if _, err := fmt.Fprintf(w, "<head>\n    "); err != nil { return err }
//line dashboard.ego:8
 _ = t.head(w) 
//line dashboard.ego:9
if _, err := fmt.Fprintf(w, "\n  "); err != nil { return err }
//line dashboard.ego:9
if _, err := fmt.Fprintf(w, "</head>\n\n  "); err != nil { return err }
//line dashboard.ego:11
if _, err := fmt.Fprintf(w, "<body class=\"index\">\n    "); err != nil { return err }
//line dashboard.ego:12
if _, err := fmt.Fprintf(w, "<div class=\"container\">\n      "); err != nil { return err }
//line dashboard.ego:13
if _, err := fmt.Fprintf(w, "<div class=\"header\">\n        "); err != nil { return err }
//line dashboard.ego:14
if _, err := fmt.Fprintf(w, "<ul class=\"nav nav-pills pull-right\">\n          "); err != nil { return err }
//line dashboard.ego:15
if _, err := fmt.Fprintf(w, "<li>"); err != nil { return err }
//line dashboard.ego:15
if _, err := fmt.Fprintf(w, "<a href=\"/_/logout\">Log out"); err != nil { return err }
//line dashboard.ego:15
if _, err := fmt.Fprintf(w, "</a>"); err != nil { return err }
//line dashboard.ego:15
if _, err := fmt.Fprintf(w, "</li>\n        "); err != nil { return err }
//line dashboard.ego:16
if _, err := fmt.Fprintf(w, "</ul>\n        "); err != nil { return err }
//line dashboard.ego:17
if _, err := fmt.Fprintf(w, "<h3 class=\"text-muted\">Gist Exposed!"); err != nil { return err }
//line dashboard.ego:17
if _, err := fmt.Fprintf(w, "</h3>\n      "); err != nil { return err }
//line dashboard.ego:18
if _, err := fmt.Fprintf(w, "</div>\n\n      "); err != nil { return err }
//line dashboard.ego:20
if _, err := fmt.Fprintf(w, "<h3>Hosted Gists"); err != nil { return err }
//line dashboard.ego:20
if _, err := fmt.Fprintf(w, "</h3>\n\n      "); err != nil { return err }
//line dashboard.ego:22
 if len(hosted) == 0 { 
//line dashboard.ego:23
if _, err := fmt.Fprintf(w, "\n        "); err != nil { return err }
//line dashboard.ego:23
if _, err := fmt.Fprintf(w, "<div class=\"row\">\n          "); err != nil { return err }
//line dashboard.ego:24
if _, err := fmt.Fprintf(w, "<div class=\"col-lg-12\">\n            "); err != nil { return err }
//line dashboard.ego:25
if _, err := fmt.Fprintf(w, "<p>You do not have any gists hosted on Gist Exposed."); err != nil { return err }
//line dashboard.ego:25
if _, err := fmt.Fprintf(w, "</p>\n          "); err != nil { return err }
//line dashboard.ego:26
if _, err := fmt.Fprintf(w, "</div>\n        "); err != nil { return err }
//line dashboard.ego:27
if _, err := fmt.Fprintf(w, "</div>\n      "); err != nil { return err }
//line dashboard.ego:28
 } else { 
//line dashboard.ego:29
if _, err := fmt.Fprintf(w, "\n        "); err != nil { return err }
//line dashboard.ego:29
if _, err := fmt.Fprintf(w, "<table class=\"table\">\n          "); err != nil { return err }
//line dashboard.ego:30
if _, err := fmt.Fprintf(w, "<thead>\n            "); err != nil { return err }
//line dashboard.ego:31
if _, err := fmt.Fprintf(w, "<tr>\n              "); err != nil { return err }
//line dashboard.ego:32
if _, err := fmt.Fprintf(w, "<th class=\"col-md-9\">Description"); err != nil { return err }
//line dashboard.ego:32
if _, err := fmt.Fprintf(w, "</th>\n              "); err != nil { return err }
//line dashboard.ego:33
if _, err := fmt.Fprintf(w, "<th class=\"col-md-3\">Created"); err != nil { return err }
//line dashboard.ego:33
if _, err := fmt.Fprintf(w, "</th>\n            "); err != nil { return err }
//line dashboard.ego:34
if _, err := fmt.Fprintf(w, "</tr>\n          "); err != nil { return err }
//line dashboard.ego:35
if _, err := fmt.Fprintf(w, "</thead>\n          "); err != nil { return err }
//line dashboard.ego:36
if _, err := fmt.Fprintf(w, "<tbody>\n            "); err != nil { return err }
//line dashboard.ego:37
 for _, g := range hosted { 
//line dashboard.ego:38
if _, err := fmt.Fprintf(w, "\n              "); err != nil { return err }
//line dashboard.ego:38
if _, err := fmt.Fprintf(w, "<tr>\n                "); err != nil { return err }
//line dashboard.ego:39
if _, err := fmt.Fprintf(w, "<td class=\"col-lg-9\">\n                  "); err != nil { return err }
//line dashboard.ego:40
if _, err := fmt.Fprintf(w, "<a href=\"/"); err != nil { return err }
//line dashboard.ego:40
if _, err := fmt.Fprintf(w, "%v", g.ID); err != nil { return err }
//line dashboard.ego:40
if _, err := fmt.Fprintf(w, "\" target=\"_blank\">\n                    "); err != nil { return err }
//line dashboard.ego:41
 if g.Description != "" { 
//line dashboard.ego:42
if _, err := fmt.Fprintf(w, "\n                      "); err != nil { return err }
//line dashboard.ego:42
if _, err := fmt.Fprintf(w, "%v",  g.Description ); err != nil { return err }
//line dashboard.ego:43
if _, err := fmt.Fprintf(w, "\n                    "); err != nil { return err }
//line dashboard.ego:43
 } else { 
//line dashboard.ego:44
if _, err := fmt.Fprintf(w, "\n                      "); err != nil { return err }
//line dashboard.ego:44
if _, err := fmt.Fprintf(w, "<em>Untitled"); err != nil { return err }
//line dashboard.ego:44
if _, err := fmt.Fprintf(w, "</em>\n                    "); err != nil { return err }
//line dashboard.ego:45
 } 
//line dashboard.ego:46
if _, err := fmt.Fprintf(w, "\n                  "); err != nil { return err }
//line dashboard.ego:46
if _, err := fmt.Fprintf(w, "</a>\n                "); err != nil { return err }
//line dashboard.ego:47
if _, err := fmt.Fprintf(w, "</td>\n                "); err != nil { return err }
//line dashboard.ego:48
if _, err := fmt.Fprintf(w, "<td class=\"col-lg-3\">\n                  "); err != nil { return err }
//line dashboard.ego:49
if _, err := fmt.Fprintf(w, "%v",  g.CreatedAt.Format(time.Stamp) ); err != nil { return err }
//line dashboard.ego:50
if _, err := fmt.Fprintf(w, "\n                "); err != nil { return err }
//line dashboard.ego:50
if _, err := fmt.Fprintf(w, "</td>\n              "); err != nil { return err }
//line dashboard.ego:51
if _, err := fmt.Fprintf(w, "</tr>\n            "); err != nil { return err }
//line dashboard.ego:52
 } 
//line dashboard.ego:53
if _, err := fmt.Fprintf(w, "\n          "); err != nil { return err }
//line dashboard.ego:53
if _, err := fmt.Fprintf(w, "</tbody>\n        "); err != nil { return err }
//line dashboard.ego:54
if _, err := fmt.Fprintf(w, "</table>\n      "); err != nil { return err }
//line dashboard.ego:55
 } 
//line dashboard.ego:56
if _, err := fmt.Fprintf(w, "\n\n\n      "); err != nil { return err }
//line dashboard.ego:58
if _, err := fmt.Fprintf(w, "<h3>Recent Gists"); err != nil { return err }
//line dashboard.ego:58
if _, err := fmt.Fprintf(w, "</h3>\n\n      "); err != nil { return err }
//line dashboard.ego:60
 if len(recent) == 0 { 
//line dashboard.ego:61
if _, err := fmt.Fprintf(w, "\n        "); err != nil { return err }
//line dashboard.ego:61
if _, err := fmt.Fprintf(w, "<div class=\"row\">\n          "); err != nil { return err }
//line dashboard.ego:62
if _, err := fmt.Fprintf(w, "<div class=\"col-lg-12\">\n            "); err != nil { return err }
//line dashboard.ego:63
if _, err := fmt.Fprintf(w, "<p>You do not have any gists available on GitHub."); err != nil { return err }
//line dashboard.ego:63
if _, err := fmt.Fprintf(w, "</p>\n          "); err != nil { return err }
//line dashboard.ego:64
if _, err := fmt.Fprintf(w, "</div>\n        "); err != nil { return err }
//line dashboard.ego:65
if _, err := fmt.Fprintf(w, "</div>\n      "); err != nil { return err }
//line dashboard.ego:66
 } else { 
//line dashboard.ego:67
if _, err := fmt.Fprintf(w, "\n        "); err != nil { return err }
//line dashboard.ego:67
if _, err := fmt.Fprintf(w, "<table class=\"table\">\n          "); err != nil { return err }
//line dashboard.ego:68
if _, err := fmt.Fprintf(w, "<thead>\n            "); err != nil { return err }
//line dashboard.ego:69
if _, err := fmt.Fprintf(w, "<tr>\n              "); err != nil { return err }
//line dashboard.ego:70
if _, err := fmt.Fprintf(w, "<th class=\"col-lg-9\">Description"); err != nil { return err }
//line dashboard.ego:70
if _, err := fmt.Fprintf(w, "</th>\n              "); err != nil { return err }
//line dashboard.ego:71
if _, err := fmt.Fprintf(w, "<th class=\"col-lg-3\">Created"); err != nil { return err }
//line dashboard.ego:71
if _, err := fmt.Fprintf(w, "</th>\n            "); err != nil { return err }
//line dashboard.ego:72
if _, err := fmt.Fprintf(w, "</tr>\n          "); err != nil { return err }
//line dashboard.ego:73
if _, err := fmt.Fprintf(w, "</thead>\n          "); err != nil { return err }
//line dashboard.ego:74
if _, err := fmt.Fprintf(w, "<tbody>\n            "); err != nil { return err }
//line dashboard.ego:75
 for _, g := range recent { 
//line dashboard.ego:76
if _, err := fmt.Fprintf(w, "\n              "); err != nil { return err }
//line dashboard.ego:76
if _, err := fmt.Fprintf(w, "<tr>\n                "); err != nil { return err }
//line dashboard.ego:77
if _, err := fmt.Fprintf(w, "<td class=\"col-md-9\">\n                  "); err != nil { return err }
//line dashboard.ego:78
if _, err := fmt.Fprintf(w, "<a href=\"/"); err != nil { return err }
//line dashboard.ego:78
if _, err := fmt.Fprintf(w, "%v", g.ID); err != nil { return err }
//line dashboard.ego:78
if _, err := fmt.Fprintf(w, "\" target=\"_blank\">\n                    "); err != nil { return err }
//line dashboard.ego:79
 if g.Description != "" { 
//line dashboard.ego:80
if _, err := fmt.Fprintf(w, "\n                      "); err != nil { return err }
//line dashboard.ego:80
if _, err := fmt.Fprintf(w, "%v",  g.Description ); err != nil { return err }
//line dashboard.ego:81
if _, err := fmt.Fprintf(w, "\n                    "); err != nil { return err }
//line dashboard.ego:81
 } else { 
//line dashboard.ego:82
if _, err := fmt.Fprintf(w, "\n                      "); err != nil { return err }
//line dashboard.ego:82
if _, err := fmt.Fprintf(w, "<em>Untitled"); err != nil { return err }
//line dashboard.ego:82
if _, err := fmt.Fprintf(w, "</em>\n                    "); err != nil { return err }
//line dashboard.ego:83
 } 
//line dashboard.ego:84
if _, err := fmt.Fprintf(w, "\n                  "); err != nil { return err }
//line dashboard.ego:84
if _, err := fmt.Fprintf(w, "</a>\n                "); err != nil { return err }
//line dashboard.ego:85
if _, err := fmt.Fprintf(w, "</td>\n                "); err != nil { return err }
//line dashboard.ego:86
if _, err := fmt.Fprintf(w, "<td class=\"col-md-3\">\n                  "); err != nil { return err }
//line dashboard.ego:87
if _, err := fmt.Fprintf(w, "%v",  g.CreatedAt.Format(time.Stamp) ); err != nil { return err }
//line dashboard.ego:88
if _, err := fmt.Fprintf(w, "\n                "); err != nil { return err }
//line dashboard.ego:88
if _, err := fmt.Fprintf(w, "</td>\n              "); err != nil { return err }
//line dashboard.ego:89
if _, err := fmt.Fprintf(w, "</tr>\n            "); err != nil { return err }
//line dashboard.ego:90
 } 
//line dashboard.ego:91
if _, err := fmt.Fprintf(w, "\n          "); err != nil { return err }
//line dashboard.ego:91
if _, err := fmt.Fprintf(w, "</tbody>\n        "); err != nil { return err }
//line dashboard.ego:92
if _, err := fmt.Fprintf(w, "</table>\n      "); err != nil { return err }
//line dashboard.ego:93
 } 
//line dashboard.ego:94
if _, err := fmt.Fprintf(w, "\n\n    "); err != nil { return err }
//line dashboard.ego:95
if _, err := fmt.Fprintf(w, "</div> "); err != nil { return err }
//line dashboard.ego:95
if _, err := fmt.Fprintf(w, "<!-- /container -->\n  "); err != nil { return err }
//line dashboard.ego:96
if _, err := fmt.Fprintf(w, "</body>\n"); err != nil { return err }
//line dashboard.ego:97
if _, err := fmt.Fprintf(w, "</html>\n\n"); err != nil { return err }
return nil
}
//line head.ego:1
 func (t *tmpl) head(w io.Writer) error  {
//line head.ego:2
if _, err := fmt.Fprintf(w, "\n\n"); err != nil { return err }
//line head.ego:3
if _, err := fmt.Fprintf(w, "<meta charset=\"utf-8\">\n"); err != nil { return err }
//line head.ego:4
if _, err := fmt.Fprintf(w, "<meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n"); err != nil { return err }
//line head.ego:5
if _, err := fmt.Fprintf(w, "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n"); err != nil { return err }
//line head.ego:6
if _, err := fmt.Fprintf(w, "<meta name=\"description\" content=\"Open source funnel analysis\">\n\n"); err != nil { return err }
//line head.ego:8
if _, err := fmt.Fprintf(w, "<title>Gist Exposed!"); err != nil { return err }
//line head.ego:8
if _, err := fmt.Fprintf(w, "</title>\n\n"); err != nil { return err }
//line head.ego:10
if _, err := fmt.Fprintf(w, "<link href=\"//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css\" rel=\"stylesheet\">\n"); err != nil { return err }
//line head.ego:11
if _, err := fmt.Fprintf(w, "<style>\n    /* Space out content a bit */\n    body {\n      padding-top: 20px;\n      padding-bottom: 20px;\n    }\n\n    .header,\n    .marketing,\n    .footer {\n      padding-right: 15px;\n      padding-left: 15px;\n    }\n\n    .header {\n      border-bottom: 1px solid #e5e5e5;\n    }\n    .header h3 {\n      padding-bottom: 19px;\n      margin-top: 0;\n      margin-bottom: 0;\n      line-height: 40px;\n    }\n\n    .footer {\n      padding-top: 19px;\n      color: #777;\n      border-top: 1px solid #e5e5e5;\n    }\n\n    @media (min-width: 768px) {\n      .container {\n        max-width: 730px;\n      }\n    }\n    .container-narrow > hr {\n      margin: 30px 0;\n    }\n\n    .jumbotron {\n      text-align: center;\n      border-bottom: 1px solid #e5e5e5;\n    }\n    .jumbotron .btn {\n      padding: 14px 24px;\n      font-size: 21px;\n    }\n\n    .marketing {\n      margin: 40px 0;\n    }\n    .marketing p + h4 {\n      margin-top: 28px;\n    }\n\n    @media screen and (min-width: 768px) {\n      .header,\n      .marketing,\n      .footer {\n        padding-right: 0;\n        padding-left: 0;\n      }\n      .header {\n        margin-bottom: 30px;\n      }\n      .jumbotron {\n        border-bottom: 0;\n      }\n    }\n"); err != nil { return err }
//line head.ego:80
if _, err := fmt.Fprintf(w, "</style>\n\n"); err != nil { return err }
//line head.ego:82
if _, err := fmt.Fprintf(w, "<script src=\"//code.jquery.com/jquery-2.1.1.min.js\">"); err != nil { return err }
//line head.ego:82
if _, err := fmt.Fprintf(w, "</script>\n"); err != nil { return err }
//line head.ego:83
if _, err := fmt.Fprintf(w, "<script src=\"//maxcdn.bootstrapcdn.com/bootstrap/3.2.0/js/bootstrap.min.js\">"); err != nil { return err }
//line head.ego:83
if _, err := fmt.Fprintf(w, "</script>\n"); err != nil { return err }
return nil
}
//line index.ego:1
 func (t *tmpl) Index(w io.Writer) error  {
//line index.ego:2
if _, err := fmt.Fprintf(w, "\n\n"); err != nil { return err }
//line index.ego:3
if _, err := fmt.Fprintf(w, "<!DOCTYPE html>\n"); err != nil { return err }
//line index.ego:4
if _, err := fmt.Fprintf(w, "<html lang=\"en\">\n  "); err != nil { return err }
//line index.ego:5
if _, err := fmt.Fprintf(w, "<head>\n    "); err != nil { return err }
//line index.ego:6
 _ = t.head(w) 
//line index.ego:7
if _, err := fmt.Fprintf(w, "\n  "); err != nil { return err }
//line index.ego:7
if _, err := fmt.Fprintf(w, "</head>\n\n  "); err != nil { return err }
//line index.ego:9
if _, err := fmt.Fprintf(w, "<body class=\"index\">\n    "); err != nil { return err }
//line index.ego:10
if _, err := fmt.Fprintf(w, "<div class=\"container\">\n      "); err != nil { return err }
//line index.ego:11
if _, err := fmt.Fprintf(w, "<div class=\"header\">\n        "); err != nil { return err }
//line index.ego:12
if _, err := fmt.Fprintf(w, "<ul class=\"nav nav-pills pull-right\">\n          "); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "<li>"); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "<a href=\"/_/login\">Sign in"); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "</a>"); err != nil { return err }
//line index.ego:13
if _, err := fmt.Fprintf(w, "</li>\n        "); err != nil { return err }
//line index.ego:14
if _, err := fmt.Fprintf(w, "</ul>\n        "); err != nil { return err }
//line index.ego:15
if _, err := fmt.Fprintf(w, "<h3 class=\"text-muted\">Gist Exposed!"); err != nil { return err }
//line index.ego:15
if _, err := fmt.Fprintf(w, "</h3>\n      "); err != nil { return err }
//line index.ego:16
if _, err := fmt.Fprintf(w, "</div>\n\n      "); err != nil { return err }
//line index.ego:18
if _, err := fmt.Fprintf(w, "<div class=\"jumbotron\">\n        "); err != nil { return err }
//line index.ego:19
if _, err := fmt.Fprintf(w, "<h1>Embed Your Gists"); err != nil { return err }
//line index.ego:19
if _, err := fmt.Fprintf(w, "</h1>\n        "); err != nil { return err }
//line index.ego:20
if _, err := fmt.Fprintf(w, "<p class=\"lead\">\n          Gist Exposed is a simple service for mirroring GitHub gists and allowing you to embed them on other sites.\n        "); err != nil { return err }
//line index.ego:22
if _, err := fmt.Fprintf(w, "</p>\n        "); err != nil { return err }
//line index.ego:23
if _, err := fmt.Fprintf(w, "<p>\n            "); err != nil { return err }
//line index.ego:24
if _, err := fmt.Fprintf(w, "<a class=\"btn btn-lg btn-success\" href=\"/_/login\" role=\"button\">Sign in with GitHub"); err != nil { return err }
//line index.ego:24
if _, err := fmt.Fprintf(w, "</a>\n        "); err != nil { return err }
//line index.ego:25
if _, err := fmt.Fprintf(w, "</p>\n      "); err != nil { return err }
//line index.ego:26
if _, err := fmt.Fprintf(w, "</div>\n\n      "); err != nil { return err }
//line index.ego:28
if _, err := fmt.Fprintf(w, "<div class=\"row marketing\">\n        "); err != nil { return err }
//line index.ego:29
if _, err := fmt.Fprintf(w, "<div class=\"col-lg-6\">\n          "); err != nil { return err }
//line index.ego:30
if _, err := fmt.Fprintf(w, "<h4>oEmbed API"); err != nil { return err }
//line index.ego:30
if _, err := fmt.Fprintf(w, "</h4>\n          "); err != nil { return err }
//line index.ego:31
if _, err := fmt.Fprintf(w, "<p>\n            Sites can use the "); err != nil { return err }
//line index.ego:32
if _, err := fmt.Fprintf(w, "<a href=\"http://oembed.com/\">oEmbed"); err != nil { return err }
//line index.ego:32
if _, err := fmt.Fprintf(w, "</a> API to create embeddable iframes to host your gists.\n          "); err != nil { return err }
//line index.ego:33
if _, err := fmt.Fprintf(w, "</p>\n        "); err != nil { return err }
//line index.ego:34
if _, err := fmt.Fprintf(w, "</div>\n\n        "); err != nil { return err }
//line index.ego:36
if _, err := fmt.Fprintf(w, "<div class=\"col-lg-6\">\n          "); err != nil { return err }
//line index.ego:37
if _, err := fmt.Fprintf(w, "<h4>Chromeless"); err != nil { return err }
//line index.ego:37
if _, err := fmt.Fprintf(w, "</h4>\n          "); err != nil { return err }
//line index.ego:38
if _, err := fmt.Fprintf(w, "<p>\n            Gists are displayed as-is with no branding or border.\n            Simply drop them into your site and style them however you'd like.\n          "); err != nil { return err }
//line index.ego:41
if _, err := fmt.Fprintf(w, "</p>\n        "); err != nil { return err }
//line index.ego:42
if _, err := fmt.Fprintf(w, "</div>\n      "); err != nil { return err }
//line index.ego:43
if _, err := fmt.Fprintf(w, "</div>\n    "); err != nil { return err }
//line index.ego:44
if _, err := fmt.Fprintf(w, "</div> "); err != nil { return err }
//line index.ego:44
if _, err := fmt.Fprintf(w, "<!-- /container -->\n  "); err != nil { return err }
//line index.ego:45
if _, err := fmt.Fprintf(w, "</body>\n"); err != nil { return err }
//line index.ego:46
if _, err := fmt.Fprintf(w, "</html>\n"); err != nil { return err }
return nil
}
