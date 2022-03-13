package coordinator

import (
	"context"
	"errors"
	"github.com/stevequadros/uploader/config"
	"github.com/stevequadros/uploader/providers"
	"github.com/stretchr/testify/require"
	"io"
	"sort"
	"testing"
)

type testUploader struct {
	name    providers.Provider
	wantErr bool
}

var _ providers.Uploader = (*testUploader)(nil)

func (u *testUploader) Upload(ctx context.Context, bucket, key string, r io.ReadSeekCloser) error {
	if u.wantErr {
		return errors.New("")
	} else {
		return nil
	}
}
func (u *testUploader) GetName() providers.Provider {
	return u.name
}

type readerSeekerCloser struct{}

var _ io.ReadSeekCloser = (*readerSeekerCloser)(nil)

func (r readerSeekerCloser) Read(n []byte) (int, error)                { return 0, nil }
func (r readerSeekerCloser) Seek(off int64, whence int) (int64, error) { return 0, nil }
func (r readerSeekerCloser) Close() error                              { return nil }

func TestCoordinator_Do(t *testing.T) {
	bucket, key := "bucket", "key"
	type fields struct {
		config    config.Config
		uploaders []providers.Uploader
	}
	type args struct {
		ctx    context.Context
		bucket string
		key    string
		reader io.ReadSeekCloser
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    DoResult
		wantErr bool
	}{
		{
			name: "returns failures and successes",
			fields: fields{
				config: config.Config{},
				uploaders: []providers.Uploader{
					&testUploader{name: "1", wantErr: false},
					&testUploader{name: "2", wantErr: false},
					&testUploader{name: "3", wantErr: false},
				},
			},
			args: args{
				ctx:    context.Background(),
				bucket: bucket,
				key:    key,
				reader: readerSeekerCloser{},
			},
			want:    DoResult{Done: []providers.Provider{"1", "2", "3"}},
			wantErr: false,
		},
		{
			name: "failures return error and list of failed if any fail",
			fields: fields{
				config: config.Config{},
				uploaders: []providers.Uploader{
					&testUploader{name: "1", wantErr: false},
					&testUploader{name: "2", wantErr: false},
					&testUploader{name: "3", wantErr: true},
				},
			},
			args: args{
				ctx:    context.Background(),
				bucket: bucket,
				key:    key,
				reader: readerSeekerCloser{},
			},
			want:    DoResult{Done: []providers.Provider{"1", "2"}, Failed: []DoError{{providers.Provider("3"), errors.New("")}}},
			wantErr: true,
		},
		{
			name: "failures return error and list of failed if all fail",
			fields: fields{
				config: config.Config{},
				uploaders: []providers.Uploader{
					&testUploader{name: "1", wantErr: true},
					&testUploader{name: "2", wantErr: true},
					&testUploader{name: "3", wantErr: true},
				},
			},
			args: args{
				ctx:    context.Background(),
				bucket: bucket,
				key:    key,
				reader: readerSeekerCloser{},
			},
			want: DoResult{
				Done: nil,
				Failed: []DoError{
					{providers.Provider("1"), errors.New("")},
					{providers.Provider("2"), errors.New("")},
					{providers.Provider("3"), errors.New("")},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Coordinator{
				config:    tt.fields.config,
				uploaders: tt.fields.uploaders,
			}
			got, err := c.Do(tt.args.ctx, tt.args.bucket, tt.args.key, tt.args.reader)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// sort for easy compare since concurrency messes with order
			sort.Slice(got.Done, func(i, j int) bool { return got.Done[i] < got.Done[j] })
			sort.Slice(got.Failed, func(i, j int) bool { return got.Failed[i].Provider < got.Failed[j].Provider })
			require.Equal(t, tt.want, got)
		})
	}
}
