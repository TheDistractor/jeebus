print("hello from TRY")
-- print(package.path)

function service(req)
  print("request:", req, req.c)
  return {"reply", req}
end
