EVO is able to parse yml files as config. By default EVO tries to reach config.yml at working directory of the executable app. If EVO was not be able to find config.yml then it tries:

For Linux:
```bash
/home/username
/var
/etc
/
```

And for Windows:
```bash
%USERNAME%
%LOCALAPPDATA%
%HOMEPATH%
%APPDATA%
%ALLUSERSPROFILE%
```

You may pass different path or file name for config.yml by passig `-c` arg to executable app.

## Configuration structure
```yaml
# App configurations
app:
  name: "Sample App" #Application Name 
  language: en-GB    #Application Lang
  static: ./httpdocs #Static files path
  session-age: 60    #Login session age

  #0 to 4
  strong-pass-level: 0 #Password difficulty

jwt:
  secret: "a.random.public.key"  #JWT Secret
  issuer: "evo"                  #JWT Issuer
  audience: ["*"]                #JWT Audience
  age: 24h                       #JWT Expiration Period
  subject: "Evo App"             #JWT Subject

# Server configurations
server:
  host: "0.0.0.0"               #Host
  port:  "80"                   #Port
  https: false                  #Use https?
  cert:  cert.pem               #ssl cert
  key:   key.pem                #ssl key
  name:  "evo"                  #Server name exposed in header
  strict-routing: false         #Care about slashes at the end of urls?
  case-sensitive: false         #Care about uppercase/lowercase urls?
  max-upload-size: 5M           #Max Upload Size
  request-id: true              #Set uinque id for each request in header
  debug: true                   #Show debug data
  recover: false                #Recover on panics

# Database credentials
database:
  enabled: true                   #Use database?
  type: "sqlite"                  #Kind of database: mssql,mysql,postgres,sqlite
  server: ""                      #Server address
  database: "database.sqlite3"    #Database name
  user: "admin"                   #Database username
  pass: "super-pedro-1980"        #Database password
  cache: true                     #Cache results? 
  cache-size: 1000                #Cache size
  debug: false                    #Debug database?
  ssl-mode: "disable"             #Database over ssl
  max-open-connections: 100       #Max db concurrent connections
  max-idle-connections: 10        #Max db idle connections
  connection-max-lifetime: 1h     #Renew connection after duration

#Log to file
log:
  writefile: true              #Write Log to file
  size: 5                      #Log size in mb
  age: 7                       #Log age in days
  level: debug                 #Log level to write on file
  path: ./logs                 #Log path

#Tweaks
tweaks:
  ballast: true                #Use ballast? https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap-26c2462549a2/
  ballast-size: 100mb          #ballast 
  processors: 0                #number of processors to occupy
  prefork: false               #use prefork https://httpd.apache.org/docs/2.4/mod/prefork.html

#Cross Origin Resource Sharing
cors:
  enabled: true                                                              #enable cors?
  allowed-origins: ["*"]                                                     #cors trusted origins
  allowed-methods: ["GET","POST","HEAD","OPTION","PUT","PATCH","DELETE"]     #cors enabled methods
  allowed-credentials: true                                                  #cors enable credentials?
  allowed-headers: ["*"]                                                     #cors enabled headers
  max-age: 0                                                                 #max preflight age

#Rate Limiter
ratelimit:
  enabled: false             #enable rate limiter?
  duration: 10               #rate limiter duration to keep data
  requests: 10               #number of requests in duration
```

## Custom configuration for apps
You may add custom configuration to the apps in to ways:
Embed configuration in default app config:
```go

type Custom struct {
	Array []string `yaml:"array"`
}

cfg := Custom{}
//load custom key from app config
evo.ParseConfig("", "custom", &cfg)
log.Print(cfg)
```

Read configuration from custom config file:
```go
//load custom key from custom file
evo.ParseConfig("custom.yml", "custom", &cfg)
log.Print(cfg)
```