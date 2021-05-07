package proxy

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/gin-gonic/gin"
)

const (
	pt1 = `point01,t1=tags10,t2=tags20 f1=11i,f2=true,f3="hello" 1602581410306591000`
	pt2 = `point02,t1=tags10,t2=tags20 f1=11i,f2=true,f3="hello" 1602581410306591000`
	ob1 = `{"source":"dk1", "status":200}`
)

func TestMain(t *testing.T) {

}
