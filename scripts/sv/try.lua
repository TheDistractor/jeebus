print("hello from TRY")

function service(req)
  print("req:", req, req.c)
  t = dbKeys("/")
  print("keys:", t[1], t[2], t[3], #t)
  print("get:", dbGet("/admin/started"))
  publish("blah", req)
  return {"reply", req}
end
