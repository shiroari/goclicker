package logger

import "testing"

func TestShouldSaveStistics(t *testing.T) {

	logger := &Logger{}

	logger.On(200, "/s?uuid=aaaaaa111", []string{}, []string{})
	logger.On(200, "/s?uuid=bbbbbb222", []string{}, []string{})
	logger.On(200, "/s?uuid=bbbbbb333", []string{}, []string{})
	logger.On(200, "/s?uuid=123", []string{}, []string{})
	logger.On(200, "", []string{}, []string{})

	stat := logger.GetStat()

	if len(stat) != 2 {
		t.Errorf("Statistics should has size 2 but size is %d", len(stat))
	}

	if stat["bbbbbb"] != 2 {
		t.Errorf("Prefix 'bbbbbb' should be found 2 times but found %d time(s)", stat["bbbbbb"])
	}

}
