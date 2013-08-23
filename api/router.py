from adapters import ActionAdapter #, ForcedAdapter, ForkeeAdapter

def route_and_handle(request):
    print "i'm routing!"
    if request.get("action"):
        a = ActionAdapter()
        a.handle()

    try:
        print request["action"]
    except:
        try:
            print request["forced"]
        except:
            print request["forkee"]


    return {"success": "success"}
    
