package classifier

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

func eventsToTsv(events []*EventWrapper) string {
	var text string
	for _, e := range events {
		text += fmt.Sprintf("%s\t%d\t%d\t%s\t%d\t%d\n",
			e.Date(),                 // date
			e.deviceId,               // device id
			e.event.Header.EventType, // event type
			e.uuid.String(),          // uuid
			e.flag,                   // flag
			len(e.event.Body.Files),  // attach_count
		)
	}
	return strings.TrimSpace(text)
}

func writeTextIntoTempFile(dir, text string) (string, error) {
	tmpFile, err := ioutil.TempFile(dir, "db-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file for saving data; %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(text); err != nil {
		return "", fmt.Errorf("failed to write data into temp file; %w", err)
	}
	return filepath.ToSlash(tmpFile.Name()), nil
}

func convertJdbcUrlToGoOrm(str, username, password string) (string, *time.Location, error) {
	jdbcUrl := strings.TrimPrefix(str, "jdbc:")
	u, err := url.Parse(jdbcUrl)
	if err != nil {
		return "", nil, err
	}

	param, err := url.ParseQuery(u.RawQuery)
	timezone := param.Get("serverTimezone")
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", nil, err
	}

	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)%s?allowAllFiles=true&charset=utf8&parseTime=true&loc=%s",
		username,
		password,
		u.Host,
		u.Path,
		strings.Replace(param.Get("serverTimezone"), "/", "%2F", -1),
	)

	return connStr, loc, nil
}
