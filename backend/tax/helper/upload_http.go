package helper

import (
	"context"
	"encoding/json"
	"net/http"

	"encore.dev/storage/objects"
)

// MaxUploadBytes caps incoming multipart payloads to 50 MiB.
const MaxUploadBytes int64 = 50 << 20

// uploadsBucket is the Encore-managed object store for raw uploads.
var uploadsBucket = objects.NewBucket("tax-uploads", objects.BucketConfig{})

// PutUploadObject streams bytes into the uploads object bucket.
func PutUploadObject(ctx context.Context, key string, data []byte) error {
	w := uploadsBucket.Upload(ctx, key)
	if _, err := w.Write(data); err != nil {
		w.Abort(err)
		return err
	}
	return w.Close()
}

// WriteJSON writes a JSON body with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteErr writes a JSON error envelope.
func WriteErr(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}
