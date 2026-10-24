package user

import (
  "bytes"
  "errors"
  "image"
  "image/png"
  "testing"

  "github.com/google/uuid"
)

func TestProcessAvatar(t *testing.T) {
  // неквадратный PNG — проверяем кавер-кроп по центру и ресайз в 256×256.
  src := image.NewRGBA(image.Rect(0, 0, 100, 50))
  var buf bytes.Buffer
  if err := png.Encode(&buf, src); err != nil {
    t.Fatalf("encode png: %v", err)
  }

  out, err := processAvatar(buf.Bytes(), 256, 82, 8000)
  if err != nil {
    t.Fatalf("process: %v", err)
  }

  img, format, err := image.Decode(bytes.NewReader(out))
  if err != nil {
    t.Fatalf("decode result: %v", err)
  }
  if format != "jpeg" {
    t.Fatalf("want jpeg, got %q", format)
  }
  if b := img.Bounds(); b.Dx() != 256 || b.Dy() != 256 {
    t.Fatalf("want 256x256, got %dx%d", b.Dx(), b.Dy())
  }
}

func TestProcessAvatarUnsupportedType(t *testing.T) {
  if _, err := processAvatar([]byte("definitely not an image"), 256, 82, 8000); !errors.Is(err, ErrUnsupportedImage) {
    t.Fatalf("want ErrUnsupportedImage, got %v", err)
  }
}

func TestProcessAvatarRejectsHugeDimensions(t *testing.T) {
  // PNG 1×10000 (по высоте) — должна отбиться защита от бомбы декомпрессии.
  src := image.NewRGBA(image.Rect(0, 0, 1, 10000))
  var buf bytes.Buffer
  if err := png.Encode(&buf, src); err != nil {
    t.Fatalf("encode png: %v", err)
  }
  if _, err := processAvatar(buf.Bytes(), 256, 82, 8000); !errors.Is(err, ErrInvalidImage) {
    t.Fatalf("want ErrInvalidImage, got %v", err)
  }
}

func TestAvatarKey(t *testing.T) {
  id := uuid.New()
  const version = "abcd1234"
  if got := AvatarKey("avatars/", id, version); got != "avatars/"+id.String()+"/"+version {
    t.Fatalf("got %q", got)
  }
  if got := AvatarKey("avatars", id, version); got != "avatars/"+id.String()+"/"+version {
    t.Fatalf("prefix without slash: got %q", got)
  }
}
