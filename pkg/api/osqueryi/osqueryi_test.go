package osqueryi

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestOsqueryi(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		query    string
		expected []map[string]string
		err      error
	}{
		{
			query: "",
			err:   errors.New("no response, is a json failure"),
		},
		{
			query: "select 1",
			err:   errors.New("this is missing a semicolon"),
		},
		{
			query: "select 1;",
			expected: []map[string]string{
				map[string]string{
					"1": "1",
				},
			},
		},
		{
			query: "select config_valid, watcher, config_hash, extensions from osquery_info;",
			expected: []map[string]string{
				map[string]string{
					"config_valid": "0",
					"config_hash":  "",
					"watcher":      "-1",
					"extensions":   "inactive",
				},
			},
		},
	}

	for _, tt := range tests {

		o, err := New()
		require.NoError(t, err, "new")

		uuid, err := o.ScheduleQuery("123", tt.query)
		if tt.err != nil {
			require.Error(t, err, "schedule query %s", tt.query)
			continue
		}

		require.NoError(t, err, "schedule")
		require.NotEmpty(t, uuid, "got uuid from scheduling")

		actual, status, err := o.FetchResults(uuid)
		require.NoError(t, err, "fetch")

		require.Equal(t, "", status)
		require.EqualValues(t, tt.expected, actual)
	}

}
