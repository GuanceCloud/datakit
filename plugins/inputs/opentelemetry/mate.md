## json 数据
```
resource_spans:{
resource:{attributes:{key:"message.type"  value:{string_value:"message-name"}} 
 attributes:{key:"service.name"  value:{string_value:"testservice"}}}  
instrumentation_library_spans:{instrumentation_library:{name:"test-tracer"}  

spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:".\xbd\x06c\x10ɫ*"  
parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  name:"Sample-0" 
 kind:SPAN_KIND_INTERNAL  start_time_unix_nano:1644312397453313100 
 end_time_unix_nano:1644312398464865900  status:{}} 

 spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:"\xd0\xf3\xe0\t\x92\xea-\xcc"  parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-1"  
\kind:SPAN_KIND_INTERNAL  start_time_unix_nano:1644312398464865900  
end_time_unix_nano:1644312399469852400 
 status:{}} 
 
spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:"\x7f<\x13\xc7L\x18\xdc\xda"  
parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-2"  kind:SPAN_KIND_INTERNAL  
start_time_unix_nano:1644312399469938800 
 end_time_unix_nano:1644312400478394400  status:{}}
  
spans:{trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:"^}\x1c\x18\x7f9>\x8e"  
parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-3"  kind:SPAN_KIND_INTERNAL  
start_time_unix_nano:1644312400478485700  
end_time_unix_nano:1644312401488931400  
status:{}}}}
```

## 组装到 span
```
trace_id:"\x94<\xdf\x00zx\x82\xe7Wy\xfe\x93\xab\x19\x95a"  
span_id:".\xbd\x06c\x10ɫ*" 
 parent_span_id:"\xa7*\x80Z#\xbeL\xf6"  
name:"Sample-0"  kind:SPAN_KIND_INTERNAL  
start_time_unix_nano:1644312397453313100  
end_time_unix_nano:1644312398464865900 
 status:{}
```

## dkspan
```
{TraceID:943cdf007a7882e75779fe93ab199561 
ParentID:a72a805a23be4cf6 
SpanID:2ebd066310c9ab2a 
Service:testservice 
Resource: 
Operation:Sample-0 
Source:opentelemetry 
SpanType: 
SourceType: 
Env: 
Project: 
Version: 
Tags:map[messagetype:string_value:"message-name" servicename:string_value:"testservice"] 
EndPoint: HTTPMethod: HTTPStatusCode: ContainerHost: 
PID: 
Start:1644312397453313100 
Duration:1011552800 
Status:STATUS_CODE_UNSET 
Content: 
SampleRate:0}
```
