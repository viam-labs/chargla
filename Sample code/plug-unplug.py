import asyncio

from viam.robot.client import RobotClient
from viam.rpc.dial import Credentials, DialOptions
from viam.components.motor import Motor
from viam.components.base import Base, Vector3

async def connect1():
    creds = Credentials(
        type='robot-location-secret',
        payload='1yq41a679kopz4qnzmba58vqcj90zc32454ssmvez9j7ssk7')
    opts = RobotClient.Options(
        refresh_interval=0,
        dial_options=DialOptions(credentials=creds)
    )
    return await RobotClient.at_address('bot1-main.47b83dmplo.viam.cloud', opts)

async def connect2():
    creds = Credentials(
        type='robot-location-secret',
        payload='1yq41a679kopz4qnzmba58vqcj90zc32454ssmvez9j7ssk7')
    opts = RobotClient.Options(
        refresh_interval=0,
        dial_options=DialOptions(credentials=creds)
    )
    return await RobotClient.at_address('pi2.47b83dmplo.viam.cloud', opts)

async def switch_charging_enabled(switch):
    await switch.set_power(1)
    await asyncio.sleep(4)
    await switch.set_power(0)
    
async def switch_charging_disabled(switch):
    await switch.set_power(-1)
    await asyncio.sleep(4)
    await switch.set_power(0)
    
async def main():
    robot1 = await connect1()
    robot2 = await connect2()

    switch = Motor.from_robot(robot2, "switch")
    insert = Motor.from_robot(robot2, "insert")
    base = Base.from_robot(robot1, "base")


    print("going to plugin")
    await plugin_once_aligned(base, switch, insert)
    print("charing for a hot minute")
    await asyncio.sleep(10)
    print("unplugging")
    await unplug(base, switch, insert)
    
    await robot1.close()
    await robot2.close()

async def plugin_once_aligned(base, switch, insert):
    print("plugin_once_aligned setup")
    await switch_charging_enabled(switch)
    await insert.set_power(-1) #back it up    
    await asyncio.sleep(30)

    print("plugin_once_aligned get to right height")
    #get insert right height
    await insert.set_power(1) # plug in charge
    await asyncio.sleep(12)
    await insert.set_power(0)

    print("plugin_once_aligned move forward")
    # start moving fowrard
    await base.set_power(Vector3(x=0, y=.25, z=0), Vector3(x=0, y=0, z=0))
    await asyncio.sleep(14)
    
    print("plugin_once_aligned insert more")
    await insert.set_power(1) # plug in charge
            
    await asyncio.sleep(30)

    print("plugin_once_aligned stopping")
    
    #stop motors
    await base.set_power(Vector3(x=0, y=0, z=0), Vector3(x=0, y=0, z=0))
    await insert.set_power(0)

async def unplug(base, switch, insert):
    await switch_charging_disabled(switch)

    await insert.set_power(-1) #back it up
    await asyncio.sleep(20)

    await base.set_power(Vector3(x=0, y=-1, z=0), Vector3(x=0, y=0, z=0))
    await asyncio.sleep(8)

    await base.set_power(Vector3(x=0, y=0, z=0), Vector3(x=0, y=0, z=0))
    await insert.set_power(0)


if __name__ == '__main__':
    asyncio.run(main())
