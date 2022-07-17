package graceful

import "github.com/sirupsen/logrus"

func Recover() {
	if err := recover(); err != nil {
		logrus.Errorf("[panic recovered]: %v", err)
	}
}
