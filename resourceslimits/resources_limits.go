package resourceslimits

import (
	"github.com/sssvip/goutil/logutil"
	"golang.org/x/time/rate"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	AlgPolling Alg = iota
	AlgRandom  Alg = iota
)

type Alg int

type ResourceLimiter struct {
	Resource interface{}
	Lim      rate.Limiter
	Used     int // 使用次数
	Fail     int // 失败次数
	Time     time.Time
}

type ResourcesSpeedLimiter struct {
	Resources []ResourceLimiter
	Pointer   int // 资源数据指针
	Len       int // 资源数量len(Resources)
	limit     int // 允许每秒取多少资源
	burst     int // 缓存最大可存多少资源
	num       int // 单次取多少资源
	alg       Alg // 取资源算法
	sync.Mutex
}

func NewResourcesSpeedLimiter(resources []interface{}, frequency string, alg Alg) *ResourcesSpeedLimiter {
	// frequency="1,10"      //每秒允许取1个资源，最多缓存10个资源，单次取1个资源
	// frequency="3"         //每秒允许取3个资源，最多缓存1个资源，单次取1个资源
	// frequency="3,3"       //每秒允许取3个资源，最多缓存3个资源，单次取1个资源
	// frequency="10,3,2"    //每秒允许取10个资源，最多缓存3个资源，单次取2个资源
	var (
		err              error
		strList          = strings.Split(frequency, ",")
		limit            = 1
		burst            = 1
		num              = 1
		resourcesLimiter []ResourceLimiter
	)
	limit, err = strconv.Atoi(strList[0])
	if err != nil {
		logutil.Error.Fatalf("frequency: %s; err: %s", frequency, err.Error())
	}
	if len(strList) > 1 {
		burst, err = strconv.Atoi(strList[1])
		if err != nil {
			logutil.Error.Fatalf("frequency: %s; err: %s", frequency, err.Error())
		}
	}
	if len(strList) > 2 {
		num, err = strconv.Atoi(strList[2])
		if err != nil {
			logutil.Error.Fatalf("frequency: %s; err: %s", frequency, err.Error())
		}
	}
	for _, resource := range resources {
		resourcesLimiter = append(resourcesLimiter, ResourceLimiter{
			Resource: resource,
			Lim:      *rate.NewLimiter(rate.Limit(limit), burst)})
	}
	logutil.Console.Printf("资源初始化Len：%d, limit: %d, burst: %d", len(resources), limit, burst)
	return &ResourcesSpeedLimiter{
		Resources: resourcesLimiter,
		Pointer:   0,
		Len:       len(resourcesLimiter),
		limit:     limit,
		burst:     burst,
		num:       num,
		alg:       alg,
	}
}

func (r *ResourcesSpeedLimiter) SetResources(resources []interface{}) {
	r.Lock()
	defer r.Unlock()
	var resourcesLimiter []ResourceLimiter
	for _, resource := range resources {
		resourcesLimiter = append(resourcesLimiter, ResourceLimiter{
			Resource: resource,
			Lim:      *rate.NewLimiter(rate.Limit(r.limit), r.burst)})
	}
	r.Resources = resourcesLimiter
	r.Len = len(resourcesLimiter)
}

func (r *ResourcesSpeedLimiter) GetResources() (ok bool, pointer int, resource interface{}) {
	r.Lock()
	defer r.Unlock()
	if r.alg == AlgRandom {
		r.Pointer = rand.Intn(r.Len)
	} else {
		r.Pointer = (r.Pointer + 1) % r.Len
	}
	ok = r.Resources[r.Pointer].Lim.AllowN(time.Now(), r.num) // 每次取 num 个资源
	if ok {
		if r.Resources[r.Pointer].Time.Before(time.Now()) {
			pointer = r.Pointer
			resource = r.Resources[r.Pointer].Resource
			return
		}
		ok = false
	}
	return
}

func (r *ResourcesSpeedLimiter) SetTime(pointer int, t time.Time) {
	r.Lock()
	defer r.Unlock()
	if pointer >= 0 && r.Len > pointer {
		r.Resources[pointer].Time = t
	}
}

func (r *ResourcesSpeedLimiter) LockResources(t time.Time) {
	r.Lock()
	defer r.Unlock()
	for i := 0; i < r.Len; i++ {
		r.Resources[i].Time = t
	}
}

func (r *ResourcesSpeedLimiter) ResetStatsAll() {
	r.Lock()
	defer r.Unlock()
	for i := 0; i < r.Len; i++ {
		r.Resources[i].Used = 0
		r.Resources[i].Fail = 0
	}
}

func (r *ResourcesSpeedLimiter) ResetStats(pointer int) {
	r.Lock()
	defer r.Unlock()
	if pointer >= 0 && pointer < r.Len {
		r.Resources[pointer].Used = 0
		r.Resources[pointer].Fail = 0
	}
}

func (r *ResourcesSpeedLimiter) UpdateStats(ok bool, pointer int) {
	r.Lock()
	defer r.Unlock()
	if pointer >= 0 && pointer < r.Len {
		r.Resources[pointer].Used++
		if ok {
			r.Resources[pointer].Fail++
		}
	}
}

func (r *ResourcesSpeedLimiter) GetResourceLimiter(pointer int) (ok bool, resourceLimiter ResourceLimiter) {
	r.Lock()
	defer r.Unlock()
	if pointer >= 0 && pointer < r.Len {
		ok = true
		resourceLimiter = r.Resources[pointer]
	}
	return
}
