package service

import (
	"bytes"
	"dd-prediction-api/pkg/apperrors"
	"image"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"image/draw"
	_ "image/jpeg"
	_ "image/png"

	"dd-prediction-api/internal/storage/repository"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	xwebp "golang.org/x/image/webp"
)

type FileService struct {
	FileRepository *repository.FileRepository
}

func NewFileService(fileRepo *repository.FileRepository) (*FileService, error) {
	return &FileService{
		FileRepository: fileRepo,
	}, nil
}

var allowedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".svg":  {},
	".png":  {},
	".webp": {},
}

const (
	duelLogoWidth  int = 128
	duelLogoHeight int = 128
)

const (
	mediaFilesDir = "storage"
	duelLogos     = "duel_logos"
)

func (s *FileService) SaveDuelLogo(
	file *multipart.FileHeader,
) (string, error) {
	return s.saveFile(duelLogos, file)
}

const maxAllowedSize = 3 * 1024 * 1024 // 3 MB

func (s *FileService) saveFile(
	dir string,
	file *multipart.FileHeader,
) (string, error) {
	if file.Size > maxAllowedSize {
		return "", apperrors.BadRequest("file exceeds max allowed size")
	}

	extension := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedImageExtensions[extension]; !ok {
		return "", apperrors.BadRequest("invalid file extension")
	}

	fileContent, err := file.Open()
	if err != nil {
		return "", apperrors.Internal("failed to open an image", err)
	}
	defer func() { _ = fileContent.Close() }()

	src, err := io.ReadAll(fileContent)
	if err != nil {
		return "", apperrors.Internal("failed to read an image", err)
	}

	src, err = resizeAndCropBytes(src, extension)
	if err != nil {
		return "", apperrors.BadRequest("failed to resize image", err)
	}
	// resizeAndCropBytes re-encodes webp as PNG (hasAlpha=true for webp in decodeAnyImage).
	// Update extension so convertToWebP uses image.Decode instead of webp.Decode on PNG bytes.
	if extension == ".webp" {
		extension = ".png"
	}

	data, err := s.convertToWebP(src, extension, false, 80)
	if err != nil {
		return "", err
	}

	fileName := uuid.New().String() + ".webp"
	savePath := filepath.Join(mediaFilesDir, dir, fileName)

	if err = s.FileRepository.Save(savePath, data); err != nil {
		return "", apperrors.Internal("failed to save an image", err)
	}

	publicURL := "/" + filepath.Join(dir, fileName)

	return publicURL, nil
}

func (s *FileService) saveFileFromBytes(dir, originalFileName string, data []byte) (string, error) {

	if len(data) > maxAllowedSize {
		return "", apperrors.BadRequest("file exceeds max allowed size")
	}

	extension := strings.ToLower(filepath.Ext(originalFileName))
	if _, ok := allowedImageExtensions[extension]; !ok {
		return "", apperrors.BadRequest("invalid file extension")
	}

	fileName := uuid.New().String() + extension
	savePath := filepath.Join(mediaFilesDir, dir, fileName)

	if err := s.FileRepository.Save(savePath, data); err != nil {
		return "", apperrors.Internal("failed to save an image", err)
	}

	publicURL := "/" + filepath.Join(dir, fileName)

	return publicURL, nil
}

func (s *FileService) RemoveDuelLogo(
	fileName string,
) error {
	return s.RemoveFile(duelLogos, fileName)
}

func (s *FileService) RemoveFile(
	dir string,
	fileName string,
) error {
	path := filepath.Join(mediaFilesDir, fileName)
	cleanPath := filepath.Clean(path)

	expectedPrefix := filepath.Join(mediaFilesDir, dir)
	if !strings.HasPrefix(cleanPath, expectedPrefix) {
		return apperrors.BadRequest("invalid path: attempting directory traversal")
	}

	if _, ok := allowedImageExtensions[filepath.Ext(fileName)]; !ok {
		return apperrors.BadRequest("invalid file extension")
	}

	if err := s.FileRepository.Remove(cleanPath); err != nil {
		if os.IsNotExist(err) {
			return apperrors.NotFound("file not found")
		}

		return apperrors.Internal("failed to delete file", err)
	}

	return nil
}

func (s *FileService) convertToWebP(data []byte, ext string, lossless bool, quality float32) ([]byte, error) {
	var (
		img image.Image
		err error
	)

	if ext == ".webp" {
		img, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			// fallback to pure-Go decoder for VP8X/extended/lossless variants
			img, err = xwebp.Decode(bytes.NewReader(data))
		}
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
	}
	if err != nil {
		return nil, apperrors.BadRequest("invalid image content", err)
	}
	if img == nil {
		return nil, apperrors.BadRequest("invalid image content")
	}

	var buf bytes.Buffer
	opts := &webp.Options{Lossless: lossless, Quality: quality}

	if err := webp.Encode(&buf, img, opts); err != nil {
		return nil, apperrors.Internal("failed to encode webp", err)
	}

	return buf.Bytes(), nil
}

func decodeAnyImage(src []byte, ext string) (image.Image, bool, error) {
	// SVG cannot be decoded as a raster image, it must be rendered to a bitmap first
	if ext == ".svg" {
		// Parse SVG data from byte stream
		icon, err := oksvg.ReadIconStream(bytes.NewReader(src))
		if err != nil {
			return nil, false, err
		}

		// Create an RGBA canvas with the target resolution (keeps transparency)
		rgba := image.NewRGBA(image.Rect(0, 0, duelLogoWidth, duelLogoHeight))

		// Map SVG viewBox to the exact output dimensions
		icon.SetTarget(0, 0, float64(duelLogoWidth), float64(duelLogoHeight))

		// Rasterize vector paths into the RGBA buffer
		scanner := rasterx.NewScannerGV(duelLogoWidth, duelLogoHeight, rgba, rgba.Bounds())
		raster := rasterx.NewDasher(duelLogoWidth, duelLogoHeight, scanner)
		icon.Draw(raster, 1.0)

		// SVGs usually contain transparency, so alpha is assumed
		return rgba, true, nil
	}

	if ext == ".webp" {
		img, err := webp.Decode(bytes.NewReader(src))
		if err != nil {
			// fallback to pure-Go decoder for VP8X/extended/lossless variants
			img, err = xwebp.Decode(bytes.NewReader(src))
			if err != nil {
				return nil, false, err
			}
		}
		return img, true, nil
	}

	// Decode PNG/JPEG into image.Image
	img, _, err := image.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, false, err
	}

	// Detect whether decoded image format supports alpha channel
	hasAlpha := false
	switch img.(type) {
	case *image.NRGBA, *image.NRGBA64,
		*image.RGBA, *image.RGBA64,
		*image.Alpha, *image.Alpha16:
		hasAlpha = true
	}

	// Convert paletted images to RGBA to avoid issues with resizing and alpha
	if _, ok := img.(*image.Paletted); ok {
		rgba := image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
		img = rgba
		hasAlpha = true
	}

	return img, hasAlpha, nil
}

func resizeAndCropBytes(src []byte, ext string) ([]byte, error) {
	// Decode image and detect alpha channel
	img, hasAlpha, err := decodeAnyImage(src, ext)
	if err != nil {
		return nil, err
	}

	// Resize image to fully cover target size and crop overflow from center
	out := imaging.Fill(
		img,
		duelLogoWidth,
		duelLogoHeight,
		imaging.Center,
		imaging.Lanczos,
	)

	buf := new(bytes.Buffer)

	// Preserve transparency by using PNG when alpha channel exists
	if hasAlpha {
		if err := imaging.Encode(buf, out, imaging.PNG); err != nil {
			return nil, err
		}
	} else {
		// Use JPEG for photos to reduce size before WebP conversion
		if err := imaging.Encode(buf, out, imaging.JPEG, imaging.JPEGQuality(90)); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
