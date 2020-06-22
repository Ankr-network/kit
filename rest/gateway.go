package rest

import (
	"context"
	"fmt"
	"github.com/Ankr-network/kit/rest/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/code"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

func CustomRESTErrorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	// return Internal when Marshal failed
	const fallback = `{"error":"InternalError", "code":13, "status":"INTERNAL", "message":"failed to marshal error message"}`

	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	w.Header().Del("Trailer")

	contentType := m.ContentType()
	// Check marshaller on run time in order to keep backwards compatibility
	// An interface param needs to be added to the ContentType() function on
	// the Marshal interface to be able to remove this check
	if httpBodyMarshaller, ok := m.(*runtime.HTTPBodyMarshaler); ok {
		pb := s.Proto()
		contentType = httpBodyMarshaller.ContentTypeFromMessage(pb)
	}
	w.Header().Set("Content-Type", contentType)

	body := fromStatus(s.Proto())
	buf, mErr := m.Marshal(body)
	if mErr != nil {
		log.Error("marshal error message error", zap.Error(mErr), zap.Reflect("body", body))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			log.Error("write fallback error body error", zap.Error(err))
		}
		return
	}

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		log.Error("extract ServerMetadata from context error")
	} else {
		handleForwardResponseServerMetadata(w, mux, md)
		handleForwardResponseTrailerHeader(w, md)
	}

	st := runtime.HTTPStatusFromCode(s.Code())
	w.WriteHeader(st)
	if _, err := w.Write(buf); err != nil {
		log.Error("write error body error", zap.Error(err))
	}

	handleForwardResponseTrailer(w, md)
}

func fromStatus(s *spb.Status) *proto.Error {
	errString := s.Message
	out := new(proto.Error)
	idx := strings.IndexRune(errString, ':')
	if idx < 0 {
		out.Error = strings.TrimSpace(errString)
		out.Message = out.Error
	} else {
		out.Error = strings.TrimSpace(errString[0:idx])
		out.Message = strings.TrimSpace(errString[idx+1:])
	}

	out.Code = s.Code
	out.Status = code.Code(s.Code)
	out.Details = s.Details
	return out
}

func handleForwardResponseServerMetadata(w http.ResponseWriter, _ *runtime.ServeMux, md runtime.ServerMetadata) {
	for k, vs := range md.HeaderMD {
		if h, ok := outgoingHeaderMatcher(k); ok {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
	}
}

func handleForwardResponseTrailerHeader(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k := range md.TrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
		w.Header().Add("Trailer", tKey)
	}
}

func handleForwardResponseTrailer(w http.ResponseWriter, md runtime.ServerMetadata) {
	for k, vs := range md.TrailerMD {
		tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
		for _, v := range vs {
			w.Header().Add(tKey, v)
		}
	}
}

func outgoingHeaderMatcher(key string) (string, bool) {
	return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
}
