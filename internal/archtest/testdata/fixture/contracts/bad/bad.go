// Package bad 故意违反「contracts 不得 import 本 module 其他模块」。
package bad

import "example.com/fixture/sdk"

var V = sdk.V
