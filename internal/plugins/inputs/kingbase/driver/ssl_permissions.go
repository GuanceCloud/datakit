//go:build !windows
// +build !windows

/*
*****************************************************************************
* 版权信息：中电科金仓（北京）科技股份有限公司

* 作者：KingbaseES

* 文件名：ssl_permissions.go

* 功能描述：检查ssl秘钥文件权限

* 其它说明：

  - 修改记录：
    1.修改时间：

    2.修改人：

    3.修改内容：

*****************************************************************************
*/
package driver

import "os"

// sslKeyPermissions检查用户提供的ssl秘钥文件的权限
func sslKeyPermissions(sslkey string) (err error) {
	info, err := os.Stat(sslkey)
	if nil != err {
		return
	}
	if 0 != info.Mode().Perm()&0077 {
		err = ErrSSLKeyHasWorldPermissions
		return
	}
	err = nil
	return
}
