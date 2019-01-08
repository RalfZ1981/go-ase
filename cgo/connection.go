package cgo

//#include <stdlib.h>
//#include "ctlib.h"
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/SAP/go-ase/libase"
)

// connection is the struct which represents a database connection.
type connection struct {
	conn *C.CS_CONNECTION
}

// newConnection allocated initializes a new connection based on the
// options in the dsn.
func newConnection(dsn libase.DsnInfo) (*connection, error) {
	err := driverCtx.init()
	if err != nil {
		return nil, fmt.Errorf("Failed to ensure context: %v", err)
	}

	conn := &connection{}

	retval := C.ct_con_alloc(driverCtx.ctx, &conn.conn)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_con_alloc failed")
	}

	// Set username.
	username := unsafe.Pointer(C.CString(dsn.Username))
	defer C.free(username)
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_USERNAME, username, C.CS_NULLTERM, nil)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_con_props failed for CS_USERNAME")
	}

	// Set password encryption
	cTrue := C.CS_TRUE
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SEC_EXTENDED_ENCRYPTION,
		unsafe.Pointer(&cTrue), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SEC_EXTENDED_ENCRYPTION")
	}

	cFalse := C.CS_FALSE
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SEC_NON_ENCRYPTION_RETRY,
		unsafe.Pointer(&cFalse), C.CS_UNUSED, nil)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SEC_NON_ENCRYPTION_RETRY")
	}

	// Set password.
	password := unsafe.Pointer(C.CString(dsn.Password))
	defer C.free(password)
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_PASSWORD, password, C.CS_NULLTERM, nil)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_con_props failed for CS_PASSWORD")
	}

	// Set hostname and port.
	hostport := unsafe.Pointer(C.CString(dsn.Host + " " + dsn.Port))
	defer C.free(hostport)
	retval = C.ct_con_props(conn.conn, C.CS_SET, C.CS_SERVERADDR, hostport, C.CS_NULLTERM, nil)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_con_props failed for CS_SERVERADDR")
	}

	retval = C.ct_connect(conn.conn, nil, 0)
	if retval != C.CS_SUCCEED {
		conn.drop()
		return nil, makeError(retval, "C.ct_connect failed")
	}

	return conn, nil
}

// drop closes and deallocates the connection.
func (conn *connection) drop() error {
	// Call context.drop when exiting this function to decrease the
	// connection counter and potentially deallocate the context.
	defer driverCtx.drop()

	retval := C.ct_close(conn.conn, C.CS_UNUSED)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_close failed, connection has results pending")
	}

	retval = C.ct_con_drop(conn.conn)
	if retval != C.CS_SUCCEED {
		return makeError(retval, "C.ct_con_drop failed")
	}

	conn.conn = nil
	return nil
}
