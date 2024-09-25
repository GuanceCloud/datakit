package def

type TypeID uint64

type TypeMap struct {
	IDMap   map[TypeID]*Class
	NameMap map[string]*Class

	T_STRING  TypeID
	T_INT     TypeID
	T_LONG    TypeID
	T_FLOAT   TypeID
	T_BOOLEAN TypeID

	T_CLASS        TypeID
	T_THREAD       TypeID
	T_FRAME_TYPE   TypeID
	T_THREAD_STATE TypeID
	T_STACK_TRACE  TypeID
	T_METHOD       TypeID
	T_PACKAGE      TypeID
	T_SYMBOL       TypeID
	T_LOG_LEVEL    TypeID

	T_STACK_FRAME  TypeID
	T_CLASS_LOADER TypeID

	T_EXECUTION_SAMPLE   TypeID
	T_ALLOC_IN_NEW_TLAB  TypeID
	T_ALLOC_OUTSIDE_TLAB TypeID
	T_LIVE_OBJECT        TypeID
	T_MONITOR_ENTER      TypeID
	T_THREAD_PARK        TypeID
	T_ACTIVE_SETTING     TypeID
}
