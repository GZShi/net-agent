package tunnel

import (
	"io"
	"reflect"
	"testing"
)

func TestServer_NewStreamRW(t *testing.T) {
	type args struct {
		SessionID uint32
	}
	tests := []struct {
		name string
		s    *Server
		args args
		want io.ReadWriter
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.NewStreamRW(tt.args.SessionID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.NewStreamRW() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_streamRW_Read(t *testing.T) {
	type args struct {
		buf []byte
	}
	tests := []struct {
		name    string
		stream  *streamRW
		args    args
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.Read(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("streamRW.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("streamRW.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_streamRW_Write(t *testing.T) {
	type args struct {
		buf []byte
	}
	tests := []struct {
		name    string
		stream  *streamRW
		args    args
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.stream.Write(tt.args.buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("streamRW.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("streamRW.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}
