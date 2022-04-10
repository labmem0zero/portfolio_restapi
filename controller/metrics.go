package controller

import (
	"fmt"
	gorillactx "github.com/gorilla/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/sys/windows"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type Metrics struct{
	statusCode *prometheus.CounterVec
	requestTime *prometheus.SummaryVec
	dbRequestTime *prometheus.SummaryVec
	processorTime *prometheus.CounterVec
	timeCPUnew float64
	timeCPUold float64
	pid uint32
}

type SYSTEM_TIMES struct {
	CreateTime syscall.Filetime
	ExitTime   syscall.Filetime
	KernelTime syscall.Filetime
	UserTime   syscall.Filetime
}

func InitiateMetrics() *Metrics {
	statusCode := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "employees_processed_ops_total",
		Help: "Общее число обработанных событий",
	}, []string{"method", "path", "status_code"})
	requestTime := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:    "employees_request_time",
		Help:    "Время, затрачиваемое на выполнение запроса к API",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"path", "request_time"})
	dbRequestTime := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "employees_db_request_time",
		Help: "Время, затрачиваемое на выполнение запроса в БД",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"path", "db_request","db_time"})
	processorTime := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "employees_processor_time_total",
		Help: "Затраченное процессорное время",
	}, []string{"processor_time"})
	metrics:=&Metrics{
		statusCode,
			requestTime,
			dbRequestTime,
			processorTime,
			0,
			0,
			0,
	}
	if runtime.GOOS == "windows" {
		metrics.pid, _ = windows.GetProcessId(windows.CurrentProcess())
		log.Println("PID текущего процесса:", metrics.pid)
		metrics.CollectCPUTime()
	}
	return metrics
}

func (metrics *Metrics)CollectCPUTime() {
	go func() {
		for {
			var times SYSTEM_TIMES
			h, _ := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, metrics.pid)
			defer windows.CloseHandle(h)
			syscall.GetProcessTimes(
				syscall.Handle(h),
				&times.CreateTime,
				&times.ExitTime,
				&times.KernelTime,
				&times.UserTime,
			)
			metrics.timeCPUnew = float64(times.KernelTime.HighDateTime)*429.4967296 + float64(times.KernelTime.LowDateTime)*1e-7
			metrics.processorTime.With(prometheus.Labels{"processor_time":"seconds"}).Add(metrics.timeCPUnew-metrics.timeCPUold)
			metrics.timeCPUold=metrics.timeCPUnew
			time.Sleep(10 * time.Second)
		}
	}()
}

func MetricizedStatus(statuscode int)bool{
	if statuscode>199&&statuscode<300{
		return true
	}
	if statuscode>399&&statuscode<600{
		return true
	}
	return false
}

func (metrics *Metrics)MiddlewareMetrics(handler http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		interceptor:=&responseWriterInterceptor{
			statusCode: http.StatusOK,
			ResponseWriter: w,
		}
		startTime:=time.Now()
		handler.ServeHTTP(interceptor,r)
		timeCost:=time.Since(startTime)
		dbRequest:=gorillactx.Get(r,"db-request")

		if dbRequest!=nil{
			metrics.dbRequestTime.With(prometheus.Labels{
				"path":r.RequestURI,"db_request":dbRequest.(string),
				"db_time":fmt.Sprintf("%v",gorillactx.Get(r,"db-time")),
			}).Observe(gorillactx.Get(r,"db-time").(float64))

			gorillactx.Delete(r, "db-request")
			gorillactx.Delete(r, "db-time")
		}

		if MetricizedStatus(interceptor.statusCode){
			metrics.statusCode.With(prometheus.Labels{
				"method": r.Method,
				"path": r.RequestURI,
				"status_code": strconv.Itoa(interceptor.statusCode),
			}).Inc()
		}

		metrics.requestTime.With(prometheus.Labels{
			"path":r.RequestURI,
			"request_time":timeCost.String(),
		}).Observe(timeCost.Seconds())
	})
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}


