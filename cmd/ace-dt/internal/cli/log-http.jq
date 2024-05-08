. | select((.msg == "HTTP Request") or (.msg == "HTTP Response")) 
| "### Round Trip " + (.requestID | tostring), .contents, "\n"