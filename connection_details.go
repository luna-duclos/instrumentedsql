package instrumentedsql

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type dbConnDetails struct {
	rawString string
	address   string
	host      string
	port      string
	dbSystem  string
	user      string
	dbName    string
}

func newDBConnDetails(connStr string) dbConnDetails {
	for _, strategy := range [...]func(string) (dbConnDetails, bool){
		parseGeneralDBDsn,
		parseMySqlDsn,
		parseSQLiteDsn,
		parsePostgreSQLKVDsn,
	} {
		if details, ok := strategy(connStr); ok {
			return details
		}
	}

	return dbConnDetails{rawString: connStr}
}

func parseGeneralDBDsn(dsn string) (dbConnDetails, bool) {
	// a valid URL may have the form of schema:path. As valid DSN strings
	// look like
	// - postgres://username:password@dbhost:port/dbname (Postgres)
	// - file::memory: (SQLite)
	// - username:password@tcp(dbhost:port)/dbname (MySQL)
	// simply parsing it will fail
	// for that reason we don't even try to parse the dsn if it doesn't contain
	// a schema followed by ://
	if !strings.Contains(dsn, "://") {
		return dbConnDetails{}, false
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return dbConnDetails{}, false
	}

	if u.Scheme == "" {
		return dbConnDetails{}, false
	}

	details := dbConnDetails{
		rawString: dsn,
		host:      u.Hostname(),
		port:      u.Port(),
		dbSystem:  u.Scheme,
	}

	if u.Path != "" {
		details.dbName = u.Path[1:]
	}

	if details.host != "" && details.port != "" {
		details.address = fmt.Sprintf("%s:%s", details.host, details.port)
	} else if details.host != "" {
		details.address = details.host
	}

	if u.User != nil {
		details.user = u.User.Username()

		// create a copy without user password
		u := cloneURL(u)
		u.User = url.User(details.user)
		details.rawString = u.String()
	}

	return details, true
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}

	u2 := new(url.URL)
	*u2 = *u

	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}

	return u2
}

var postgresKVPasswordRegex = regexp.MustCompile(`(^|\s)password=[^\s]+(\s|$)`)

func parsePostgreSQLKVDsn(dsn string) (dbConnDetails, bool) {
	var details dbConnDetails

	for _, kv := range strings.Split(dsn, " ") {
		field := strings.ToLower(kv)

		var (
			prefix    string
			detailPtr *string
		)
		switch {
		case strings.HasPrefix(field, "host="):
			if details.host != "" {
				// hostaddr= takes precedence
				continue
			}

			prefix, detailPtr = "host=", &details.host
		case strings.HasPrefix(field, "hostaddr="):
			prefix, detailPtr = "hostaddr=", &details.host
		case strings.HasPrefix(field, "port="):
			prefix, detailPtr = "port=", &details.port
		case strings.HasPrefix(field, "user="):
			prefix, detailPtr = "user=", &details.user
		case strings.HasPrefix(field, "dbname="):
			prefix, detailPtr = "dbname=", &details.dbName
		default:
			continue
		}

		*detailPtr = kv[len(prefix):]
	}

	if details.dbName == "" {
		return dbConnDetails{}, false
	}

	details.dbSystem = "postgres"
	details.rawString = postgresKVPasswordRegex.ReplaceAllString(dsn, " ")
	if details.host != "" && details.port != "" {
		details.address = details.host + ":" + details.port
	} else if details.host != "" {
		details.address = details.host
	}

	return details, true
}

func parseSQLiteDsn(dsn string) (dbConnDetails, bool) {
	if !strings.HasPrefix(dsn, "file:") && !strings.HasPrefix(dsn, ":memory:") {
		return dbConnDetails{}, false
	}

	details := dbConnDetails{
		rawString: dsn,
		dbSystem:  "sqlite",
	}

	if pos := strings.IndexRune(dsn, '?'); pos >= 1 {
		if params, err := url.ParseQuery(dsn[pos+1:]); err == nil {
			details.user = params.Get("_auth_user")
			if password := params.Get("_auth_pass"); password != "" {
				details.rawString = strings.Replace(details.rawString, "_auth_pass="+password, "_auth_pass=*****", -1)
			}
		}
		details.address = dsn[:pos]
	} else {
		details.address = dsn
	}

	return details, true
}

func parseMySqlDsn(dsn string) (dbConnDetails, bool) {
	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// Find the last '/' (since the password or the net addr might contain a '/')
	details := dbConnDetails{ rawString: dsn }

	foundSlash := false
	for i := len(dsn) - 1; i >= 0; i-- {
		if dsn[i] == '/' {
			foundSlash = true
			var j, k int

			// left part is empty if i <= 0
			if i > 0 {
				// [username[:password]@][protocol[(address)]]
				// Find the last '@' in dsn[:i]
				for j = i; j >= 0; j-- {
					if dsn[j] == '@' {
						// username[:password]
						// Find the first ':' in dsn[:j]
						for k = 0; k < j; k++ {
							if dsn[k] == ':' {
								password := dsn[k+1 : j]
								details.rawString = strings.Replace(details.rawString, ":"+password, ":*****", -1)
								break
							}
						}
						details.user = dsn[:k]

						break
					}
				}

				// [protocol[(address)]]
				// Find the first '(' in dsn[j+1:i]
				for k = j + 1; k < i; k++ {
					if dsn[k] == '(' {
						// dsn[i-1] must be == ')' if an address is specified
						if dsn[i-1] != ')' {
							if strings.ContainsRune(dsn[k+1:i], ')') {
								return dbConnDetails{}, false
							}
							return dbConnDetails{}, false
						}
						details.address = dsn[k+1 : i-1]
						break
					}
				}
			}

			// dbname[?param1=value1&...&paramN=valueN]
			details.dbName = dsn[i + 1:]
			if iq := strings.IndexRune(details.dbName, '?'); iq != -1 {
				details.dbName = details.dbName[:iq]
			}

			break
		}
	}

	if !foundSlash && len(dsn) > 0 {
		return dbConnDetails{}, false
	}

	details.dbSystem = "mysql"
	return details, true
}
