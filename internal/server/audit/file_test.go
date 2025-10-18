package audit

import (
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/EshkinKot1980/metrics/internal/server/audit/mocks"
)

func TestNewFileAuditor(t *testing.T) {
	tmpdir := t.TempDir()

	err := os.MkdirAll(tmpdir+"/readOnlyDir", 0555)
	require.Nil(t, err, "Read only dir creating")

	file, err := os.OpenFile(tmpdir+"/readOnlyFile", os.O_CREATE, 0444)
	require.Nil(t, err, "Read only dir creating")
	file.Close()

	tests := []struct {
		name     string
		filePath string
		err      error
	}{
		{
			name:     "susses_new_file",
			filePath: tmpdir + "/subDir/audit.log",
			err:      nil,
		},
		{
			name:     "susses_new_file_existing_dir",
			filePath: tmpdir + "/subDir/audit1.log",
			err:      nil,
		},
		{
			name:     "susses_existing_file",
			filePath: tmpdir + "/subDir/audit.log",
			err:      nil,
		},
		{
			name:     "not_directory_in_path",
			filePath: tmpdir + "/subDir/audit.log/audit.log",
			err:      errors.New("invalid file path, stat " + tmpdir + "/subDir/audit.log/audit.log: not a directory"),
		},
		{
			name:     "read_only_dir",
			filePath: tmpdir + "/readOnlyDir/audit.log",
			err:      errors.New("failed to create or open file, open " + tmpdir + "/readOnlyDir/audit.log: permission denied"),
		},
		{
			name:     "read_only_dir_make_subdir",
			filePath: tmpdir + "/readOnlyDir/subDir/audit.log",
			err:      errors.New("invalid file path, mkdir " + tmpdir + "/readOnlyDir/subDir: permission denied"),
		},
		{
			name:     "read_only_file",
			filePath: tmpdir + "/readOnlyFile",
			err:      errors.New("failed to create or open file, open " + tmpdir + "/readOnlyFile: permission denied"),
		},
		{
			name:     "directory_instead_of_file",
			filePath: tmpdir + "/subDir",
			err:      errors.New("file " + tmpdir + "/subDir already exists and it is a directory"),
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocks.NewMockLogger(ctrl)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFileAuditor(test.filePath, logger)

			if test.err == nil {
				if err != nil {
					t.Errorf(`error expected to be: "%v"; got: "%v"`, nil, err)
				}
			} else if err == nil || test.err.Error() != err.Error() {
				t.Errorf(`error expected to be: "%v"; got: "%v"`, test.err, err)
			}

		})
	}
}
