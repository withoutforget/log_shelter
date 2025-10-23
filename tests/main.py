import asyncio
import datetime
import nats
"""
Here will be integration tests
"""

raw = """
    {{
    "raw_log": "User login failed due to invalid credentials",
    "log_level": "ERROR",
    "source": "auth-service",
    "created_at": "{}",
    "request_id": "req-7a3b9c2d-1e4f-5a6b-8c9d-0e1f2a3b4c5d",
    "logger_name": "AuthServiceLogger"
    }}
""".format(datetime.datetime.now(tz = datetime.UTC).isoformat()).encode()

async def main_():    
    nc = await nats.connect(servers = "nats://localhost:4222", user = "nats", password = "nats")

    await nc.publish("nats.hi", payload = raw)

    res = await nc.request("nats.bye", payload = b"""{  "page": 1,  "levels": ["*"],  "sources": ["*"], "order": "asc"}""")
    print(res)

async def main():    
    nc = await nats.connect(servers = "nats://localhost:4222", user = "nats", password = "nats")

    data = [nc.publish("nats.hi", payload = raw) for _ in range(20)]
    await asyncio.gather(*data)

if __name__ == '__main__':
    asyncio.run(main())
