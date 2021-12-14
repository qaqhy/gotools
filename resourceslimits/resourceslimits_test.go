package resourceslimits

import (
	"github.com/sssvip/goutil/logutil"
	"testing"
	"time"
)

func TestGetResources(t *testing.T) {
	var results = []interface{}{7.999, 2, "A"}
	//lim := NewResourcesSpeedLimiter(results, "0.1,10,1", AlgPolling)  // 10秒生产1个资源，最多缓存10个资源，一次取1个资源
	lim := NewResourcesSpeedLimiter(results, "100,2,2", AlgPolling)  // 10毫秒生产1个资源，最多缓存2个资源，一次取2个资源
	lim.SetTime(1, time.Now().Add(2*time.Millisecond))  // 资源1先休息2毫秒
	for i := 0; i < 300; i++ {
		logutil.Console.Println(lim.GetResources())
		time.Sleep(time.Millisecond * 3)
	}
}