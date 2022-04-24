module use_rtmp

go 1.18

replace github.com/hsnks100/librtmp => ./librtmp

require (
	github.com/hsnks100/librtmp v0.0.0-00010101000000-000000000000
	github.com/hsnks100/tcpengine v0.0.0-20220423001146-aac43ba24768
	github.com/sirupsen/logrus v1.8.1
	github.com/yutopp/go-amf0 v0.0.0-20180803120851-48851794bb1f
)

require (
	github.com/icza/bitio v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
)
