# API Changelog 1.0.0 vs. 1.0.1

## GET /api/{domain}/{project}/badges/security-score
- :warning: removed the success response with the status '200'
- :warning: removed the success response with the status '201'
- :warning: deleted the 'cookie' request parameter 'test'
- :warning: deleted the 'header' request parameter 'user'
- :warning: deleted the 'query' request parameter 'filter'
-  api operation id 'GetSecurityScores' removed and replaced with 'GetSecurityScore'
-  api tag 'security' removed
-  for the 'query' request parameter 'token', the maxLength was increased from '29' to '30'
-  removed the pattern '^(?:[\w-./:]+)$' from the 'query' request parameter 'image'
-  for the 'query' request parameter 'image', the type/format was generalized from 'string'/'general string' to ''/''
-  removed the non-success response with the status '400'


## GET /api/{domain}/{project}/install-command
- :warning: deleted the 'header' request parameter 'network-policies'
-  added the new optional 'header' request parameter 'name' to all path's operations
-  added the new enum value 'test1' to the 'path' request parameter 'project'


## POST /register
-  the endpoint scheme security 'bearerAuth' was removed from the API
-  the security scope 'write:pets' was added to the endpoint's security scheme 'OAuth'