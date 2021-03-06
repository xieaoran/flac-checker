package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/xieaoran/flac-checker/local"
	"github.com/xieaoran/flac-checker/models"
	"github.com/xieaoran/flac-checker/remote"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getLogger(logPath string) *zap.SugaredLogger {
	logFile, _ := os.Create(logPath)

	consoleConfig := zap.NewProductionEncoderConfig()
	consoleConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)
	fileConfig := zap.NewProductionEncoderConfig()
	fileConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	fileConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewConsoleEncoder(fileConfig)
	loggerCore := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(zap.WarnLevel)),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zap.NewAtomicLevelAt(zap.DebugLevel)))

	logger := zap.New(loggerCore)
	return logger.Sugar()
}

func main() {
	rootPath := flag.String("p", "", "Path Containing FLAC Files To Be Scanned")
	logPath := flag.String("l", "result.log", "Path Of Output Log File")
	flag.Parse()

	logger := getLogger(*logPath)
	defer logger.Sync()

	metaRegistry := local.NewMetaRegistry()
	remoteProvider := remote.NewMORAProvider()

	var fileMetas []*models.FileMeta
	pathError := filepath.WalkDir(*rootPath, func(path string, d fs.DirEntry, err error) error {
		fileMeta, metaOK, metaError := metaRegistry.ParseFile(path)
		if !metaOK {
			logger.Debugf("unsupported path, skipping, path[%s], metaOK[%t]", path, metaOK)
			return nil
		}
		if metaError != nil {
			return fmt.Errorf("metaParser.ParseFile failed, "+
				"path[%s], metaError.Error[%s]", path, metaError.Error())
		}

		fileMetas = append(fileMetas, fileMeta)
		return nil
	})
	if pathError != nil {
		logger.Fatalf("filepath.WalkDir failed, "+
			"rootPath[%s], pathError.Error[%s]", *rootPath, pathError.Error())
	}

	for _, fileMeta := range fileMetas {
		remoteFiles, remoteError := remoteProvider.Search(fileMeta)
		if remoteError != nil {
			logger.Fatalf("remoteProvider.Search failed, "+
				"fileMeta[%+v], remoteError.Error[%s]", fileMeta, remoteError.Error())
		}
		if len(remoteFiles) == 0 {
			logger.Debugf("[????????????] %s", fileMeta.FilePath)
			logger.Debugf("[????????????] %s", fileMeta.Title)
			logger.Debugf("[???????????????] %s", fileMeta.Artist)
			logger.Debugf("[???????????????] %s", fileMeta.Album)
			logger.Debugf("[???????????????] %d[bit]", fileMeta.BitDepth)
			logger.Debugf("[???????????????] %d[Hz]", fileMeta.SampleRate)
			logger.Debugf("[???????????? MD5] %s", fileMeta.AudioMD5)
			logger.Debug("---------------------------------------")
			logger.Debug("????????????????????????????????????")
			logger.Debug("=======================================")
			continue
		}
		logger.Warnf("[????????????] %s", fileMeta.FilePath)
		logger.Warnf("[????????????] %s", fileMeta.Title)
		logger.Warnf("[???????????????] %s", fileMeta.Artist)
		logger.Warnf("[???????????????] %s", fileMeta.Album)
		logger.Warnf("[???????????????] %d[bit]", fileMeta.BitDepth)
		logger.Warnf("[???????????????] %d[Hz]", fileMeta.SampleRate)
		logger.Warnf("[???????????? MD5] %s", fileMeta.AudioMD5)
		logger.Warn("---------------------------------------")
		logger.Warnf("????????? [%d] ?????????????????????", len(remoteFiles))
		for _, remoteFile := range remoteFiles {
			logger.Warn("---------------------------------------")
			logger.Warnf("[????????????] %s", remoteFile.FilePath)
			logger.Warnf("[????????????] %s", remoteFile.Title)
			logger.Warnf("[???????????????] %s", remoteFile.Artist)
			logger.Warnf("[???????????????] %s", remoteFile.Album)
			logger.Warnf("[???????????????] %d[bit]", remoteFile.BitDepth)
			logger.Warnf("[???????????????] %d[Hz]", remoteFile.SampleRate)
		}
		logger.Warn("=======================================")
	}
}
