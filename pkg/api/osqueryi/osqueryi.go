package osqueryi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/AbGuthrie/goquery/hosts"
	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type osqueryResult struct {
	rows   []map[string]string
	stderr string
	err    error
}

type OsqueryiApi struct {
	results  map[string]*osqueryResult
	logger   log.Logger
	timeout  time.Duration
	osqueryd string
}

type Opt func(*OsqueryiApi) error

func WithLogger(logger log.Logger) Opt {
	return func(o *OsqueryiApi) error {
		o.logger = logger
		return nil
	}
}

func New(opts ...Opt) (*OsqueryiApi, error) {
	o := &OsqueryiApi{
		results:  make(map[string]*osqueryResult),
		logger:   log.NewNopLogger(),
		timeout:  10 * time.Second,
		osqueryd: "osqueryd",
	}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, errors.Wrap(err, "creating new osqueryi")
		}
	}

	return o, nil

}

func (o *OsqueryiApi) CheckHost(host string) (hosts.Host, error) {
	if host != "local" {
		return hosts.Host{}, errors.New(`Only supported host is "local"`)
	}

	return hosts.Host{
		UUID:             "local",
		ComputerName:     "osqueryi",
		Platform:         "local", //FIXME
		Version:          "local", //FIXME
		CurrentDirectory: "/",
	}, nil
}

func (o *OsqueryiApi) ScheduleQuery(origUuid string, query string) (string, error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), o.timeout)
	defer ctxCancel()

	uuidTyped, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "uuid creation")
	}
	uuid := uuidTyped.String()

	args := []string{
		"-S",
		"--json",
	}

	cmd := exec.CommandContext(ctx, o.osqueryd, args...)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	cmd.Stdin = strings.NewReader(query)

	cmdErr := cmd.Run()

	result := &osqueryResult{}
	o.results[uuid] = result

	stderrResult := string(stderr.Bytes())
	if cmdErr != nil || len(stderrResult) != 0 {
		result.err = errors.Errorf("osqueryi error: %s", string(stderr.Bytes()))
		return uuid, result.err
	}

	if err := json.Unmarshal(stdout.Bytes(), &result.rows); err != nil {
		fmt.Println(stdout.String())
		result.err = errors.Wrap(err, "json unmarshal")
		return uuid, result.err
	}

	return uuid, nil

}

func (o *OsqueryiApi) FetchResults(uuid string) ([]map[string]string, string, error) {
	results, ok := o.results[uuid]
	if !ok {
		return nil, "", errors.New("query not found")
	}

	return results.rows, "", results.err
}

func (o *OsqueryiApi) Close() error {
	return nil
}
