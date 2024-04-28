 
import config
from machine import Pin, Timer
import time
import network
import time

####### SETUP #########

rpi_led = Pin('LED', Pin.OUT)

bzr_active_1 = Pin(2, Pin.OUT)
bzr_active_2 = Pin(3, Pin.OUT)
bzr_active_3 = Pin(4, Pin.OUT)

bzr_active_1.value(0)
bzr_active_2.value(0)
bzr_active_3.value(0)

bzr_passive_1 = Pin(5, Pin.OUT)

bzr_passive_1.value(0)

key = Pin(10, Pin.OPEN_DRAIN)
key.value(1)

led_r = Pin(18, Pin.OUT)
led_g = Pin(19, Pin.OUT)
led_b = Pin(20, Pin.OUT)

led_r.value(0)
led_g.value(0)
led_b.value(0)

wlan = network.WLAN(network.STA_IF)
wlan.active(True)

timer_wifi = Timer()

curr_wifi_idx = 0
last_wifi_status = None

#######  #########

def wifi_connect():
    global curr_wifi_idx
    conf = config.WIFI_CONF[curr_wifi_idx]
    print('Connect to WLAN[%s]: [%s]::[%s]' % (curr_wifi_idx, conf["ssid"], conf["password"]) )
    wlan.connect(conf["ssid"], conf["password"])


wifi_connect()


def led_blink(led, times=1, delayOn=0.2, delayOff=0.2):
    led.value(0)
    for i in range(times):
        led.value(1)
        time.sleep(delayOn)
        if i < times - 1:
            led.value(0)
            time.sleep(delayOff)
    led.value(0)

def wifi_worker(timer):
    global curr_wifi_idx
    global last_wifi_status

    led_r.value(0)
    led_g.value(0)
    led_b.value(0)

    wlan_status = wlan.status()
    reconnect = False

    if wlan_status == network.STAT_CONNECTING:
        led_blink(led_b, 3)
        print('[WLAN-STATUS]: Connecting')

    elif wlan_status == network.STAT_GOT_IP:
        if last_wifi_status != wlan_status:
            led_blink(led_g, 1)
        print('[WLAN-STATUS]: Okay | Got IP')

    elif wlan_status == network.STAT_NO_AP_FOUND:
        led_blink(led_r, 1, 0.5)
        print('[WLAN-STATUS]: No AP Found')
        reconnect = True

    elif wlan_status == network.STAT_IDLE:
        led_blink(led_r, 2)
        print('[WLAN-STATUS]: Idle | No Connection')
        reconnect = True

    elif wlan_status == network.STAT_WRONG_PASSWORD:
        led_blink(led_r, 3, 0.2)
        print('[WLAN-STATUS]: Wrong PW')
        reconnect = True

    elif wlan_status == network.STAT_CONNECT_FAIL:
        led_blink(led_r, 4, 0.2)
        print('[WLAN-STATUS]: Connection Failed')
        reconnect = True

    else:
        led_blink(led_r, 10, 0.1)
        print('[WLAN-STATUS]: ?????')

    last_wifi_status = wlan_status

    if reconnect:
        curr_wifi_idx = (curr_wifi_idx + 1) % len(config.WIFI_CONF)
        wifi_connect()

    led_r.value(0)
    led_g.value(0)
    led_b.value(0)


timer_wifi.init(period=3000, mode=Timer.PERIODIC, callback=wifi_worker)

# 
# print('Start')
# 
# key.value(0)
# time.sleep(0.2)
# key.value(1)
# 
# 
# print('Okay')
# 
# 
# sleep(2)
# 
# print('Kill')
# 
# key.value(0)
# sleep(2)
# key.value(1)
# 
# print('End')


while True:
    time.sleep(1)