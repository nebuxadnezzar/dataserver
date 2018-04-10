
print( "\n\nLUA QUERY PARAMS" )
for key,value in pairs( requestParams ) do print( string.format( "%s -> %s\n", key,value) ) end

print(table.concat(requestParams, "; "))

sendData( '{"status":"ok", "handler":"printWebArgs.lua"}' )