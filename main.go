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
			logger.Debugf("[本地路径] %s", fileMeta.FilePath)
			logger.Debugf("[本地标题] %s", fileMeta.Title)
			logger.Debugf("[本地艺术家] %s", fileMeta.Artist)
			logger.Debugf("[本地唱片集] %s", fileMeta.Album)
			logger.Debugf("[本地位深度] %d[bit]", fileMeta.BitDepth)
			logger.Debugf("[本地采样率] %d[Hz]", fileMeta.SampleRate)
			logger.Debugf("[本地音频 MD5] %s", fileMeta.AudioMD5)
			logger.Debug("---------------------------------------")
			logger.Debug("未发现合适的远程提升样品")
			logger.Debug("=======================================")
			continue
		}
		logger.Warnf("[本地路径] %s", fileMeta.FilePath)
		logger.Warnf("[本地标题] %s", fileMeta.Title)
		logger.Warnf("[本地艺术家] %s", fileMeta.Artist)
		logger.Warnf("[本地唱片集] %s", fileMeta.Album)
		logger.Warnf("[本地位深度] %d[bit]", fileMeta.BitDepth)
		logger.Warnf("[本地采样率] %d[Hz]", fileMeta.SampleRate)
		logger.Warnf("[本地音频 MD5] %s", fileMeta.AudioMD5)
		logger.Warn("---------------------------------------")
		logger.Warnf("共发现 [%d] 个远程提升样品", len(remoteFiles))
		for _, remoteFile := range remoteFiles {
			logger.Warn("---------------------------------------")
			logger.Warnf("[远程路径] %s", remoteFile.FilePath)
			logger.Warnf("[远程标题] %s", remoteFile.Title)
			logger.Warnf("[远程艺术家] %s", remoteFile.Artist)
			logger.Warnf("[远程唱片集] %s", remoteFile.Album)
			logger.Warnf("[远程位深度] %d[bit]", remoteFile.BitDepth)
			logger.Warnf("[远程采样率] %d[Hz]", remoteFile.SampleRate)
		}
		logger.Warn("=======================================")
	}
}
