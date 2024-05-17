 
import config
from machine import PWM, Pin, Timer
import time
import network
import time
import urequests
import gc
import sys
import math

####### ####### ####### #######

CMD_RESET        = 0
CMD_KEY_ON       = 1
CMD_KEY_OFF      = 2
CMD_BZR1_ON      = 3
CMD_BZR1_OFF     = 4
CMD_BZR2_ON      = 5
CMD_BZR2_OFF     = 6
CMD_BZR3_ON      = 7
CMD_BZR3_OFF     = 8
CMD_PWM_ON       = 9
CMD_PWM_OFF      = 10
CMD_PWM_FUNC_ON  = 11
CMD_PWM_FUNC_OFF = 12
CMD_PWM_NOTE_ON  = 13
CMD_PWM_NOTE     = 14
CMD_PWM_NOTE_OFF = 15

####### SETUP #########

rpi_led = Pin('LED', Pin.OUT)

bzr_active_1 = Pin(2, Pin.OUT)
bzr_active_2 = Pin(3, Pin.OUT)
bzr_active_3 = Pin(4, Pin.OUT)

bzr_active_1.value(0)
bzr_active_2.value(0)
bzr_active_3.value(0)

bzr_passive_duty = int(65535 / 2) # 50% on | 50% off

bzr_passive_1 = PWM(Pin(5))
bzr_passive_1.freq(2_000)
bzr_passive_1.duty_u16(0)
bzr_passive_1.init()

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

buzz_fn_worker = Timer()

curr_wifi_idx = 0
last_wifi_status = None

pwm_func_start   = 0
pwm_func_period  = 0
pwm_func_freqmin = 0
pwm_func_freqmax = 0

cmd_queue = []

################  ################  ################  ################  ################  ################  ##################



def wifi_connect():
    global curr_wifi_idx
    conf = config.WIFI_CONF[curr_wifi_idx]
    print('Connect to WLAN[%s]: [%s]' % (curr_wifi_idx, conf["ssid"]) )
    wlan.connect(conf["ssid"], conf["password"])



def led_blink(led, times=1, delayOn=0.2, delayOff=0.2):
    led.value(0)
    for i in range(times):
        led.value(1)
        time.sleep(delayOn)
        if i < times - 1:
            led.value(0)
            time.sleep(delayOff)
    led.value(0)



def wifi_worker():
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
            print('[WLAN-STATUS]: Okay | Got IP: ' + wlan.ifconfig()[0])
            
        #print('[DBG] [WLAN-STATUS]: Still Okay')

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



def cmd_query_worker():
    global cmd_queue
    
    if not wlan.isconnected():
        print('[DBG] [No Query | Not Connected]...')
        return
    
    if len(cmd_queue) > 0 and cmd_queue[0][0] < (time.ticks_ms() + 2_000):
        print('[DBG] [No Query | Active cmd (sub 2s)]...')
        return
        
    
    #print('[DBG] [Query] ' + config.URL_QUERY)

    try:
        gc.collect()
        response = urequests.request("POST", config.URL_QUERY, headers={"Authorization": "Secret " + config.CC_SECRET})
        data = response.text
        gc.collect()
        for line in data.splitlines():
            enqueue_cmd(line)
    except Exception as e:
        print('Error while querying server: ' + str(e))
        sys.print_exception(e)
        led_blink(led_r, 2, 0.1, 0.1)
        time.sleep_ms(5)
        return
        

def enqueue_cmd(cmdstr):
    global cmd_queue
    
    split = cmdstr.split(';')
    
    print('[QQQ] enqueue_cmd: ' + cmdstr)
    
    cmd = {'command': split[1]}
    for attr in split[2:]:
        asplit = attr.split('=')
        cmd[asplit[0]] = asplit[1]
    
    t0 = time.ticks_ms() + int(cmd['delay'])
    
    if split[1] == "RESET":
        cmd_queue.append([t0, CMD_RESET, cmd])
    elif split[1] == "LAMP":
        cmd_queue.append([t0,                             CMD_KEY_ON,       cmd])
        cmd_queue.append([t0+200,                         CMD_KEY_OFF,      cmd])
        cmd_queue.append([t0 + int(cmd['duration']),      CMD_KEY_ON,       cmd])
        cmd_queue.append([t0 + int(cmd['duration'])+2000, CMD_KEY_OFF,      cmd])
    elif split[1] == "BUZZER_1":
        cmd_queue.append([t0,                             CMD_BZR1_ON,      cmd])
        cmd_queue.append([t0 + int(cmd['duration']),      CMD_BZR1_OFF,     cmd])
    elif split[1] == "BUZZER_2":
        cmd_queue.append([t0,                             CMD_BZR2_ON,      cmd])
        cmd_queue.append([t0 + int(cmd['duration']),      CMD_BZR2_OFF,     cmd])
    elif split[1] == "BUZZER_3":
        cmd_queue.append([t0,                             CMD_BZR3_ON,      cmd])
        cmd_queue.append([t0 + int(cmd['duration']),      CMD_BZR3_OFF,     cmd])
    elif split[1] == "BUZZER_PWM":
        cmd_queue.append([t0,                             CMD_PWM_ON,       cmd])
        cmd_queue.append([t0 + int(cmd['duration']),      CMD_PWM_OFF,      cmd])
    elif split[1] == "BUZZER_PWM_FUNC":
        cmd_queue.append([t0,                             CMD_PWM_FUNC_ON,  cmd])
        cmd_queue.append([t0 + int(cmd['duration']),      CMD_PWM_FUNC_OFF, cmd])
    elif split[1] == "BUZZER_PWM_NOTES":
        notelen = int(cmd['note_length'])
        tstart = t0
        needs_on = True
        for note in cmd['notes'].split(':'):
            f_note = int(note)
            if f_note == 0:
                cmd_queue.append([tstart, CMD_PWM_NOTE_OFF, {}])
                needs_on = True
            else:
                cmd_queue.append([tstart, CMD_PWM_NOTE_ON if needs_on else CMD_PWM_NOTE, {'frequency': f_note}])
            tstart += notelen
        cmd_queue.append([tstart, CMD_PWM_NOTE_OFF, {'frequency': int(note)}])
    else:
        raise ValueError("unknown cmd: " + cmdstr)
    
    cmd_queue = sorted(cmd_queue)
    
    #print('[DBG] QUEUE := ' + str(cmd_queue))


def pwm_func_sinus(timer):
    global pwm_func_start
    global pwm_func_period
    global pwm_func_freqmin
    global pwm_func_freqmax
    
    tns = time.time_ns()
    
    delta = float(tns - pwm_func_start)
    progr = (delta / float(pwm_func_period)) % 1
    val   = (math.sin(progr * math.pi * 2 - math.pi/2) + 1) / 2
    freq  = pwm_func_freqmin + (pwm_func_freqmax - pwm_func_freqmin)*val
    
    bzr_passive_1.freq(int(freq))
        
def pwm_func_triangle(timer):
    global pwm_func_start
    global pwm_func_period
    global pwm_func_freqmin
    global pwm_func_freqmax
    
    tns = time.time_ns()
    
    delta = float(tns - pwm_func_start)
    progr = (delta / float(pwm_func_period)) % 1
    if progr < 0.5:
        val = (progr / 0.5)
        freq = pwm_func_freqmin + (pwm_func_freqmax - pwm_func_freqmin)*val
        bzr_passive_1.freq(int(freq))
    else:
        val = 2 - (progr / 0.5)
        freq = pwm_func_freqmin + (pwm_func_freqmax - pwm_func_freqmin)*val
        bzr_passive_1.freq(int(freq))
    
    
def pwm_func_sawtooth(timer):
    global pwm_func_start
    global pwm_func_period
    global pwm_func_freqmin
    global pwm_func_freqmax
    
    tns = time.time_ns()
    
    delta = float(tns - pwm_func_start)
    progr = (delta / float(pwm_func_period)) % 1
    val = progr
    freq = pwm_func_freqmin + (pwm_func_freqmax - pwm_func_freqmin)*val
    bzr_passive_1.freq(int(freq))
    
    
def pwm_func_square(timer):
    global pwm_func_start
    global pwm_func_period
    global pwm_func_freqmin
    global pwm_func_freqmax
    
    tns = time.time_ns()
    
    delta = float(tns - pwm_func_start)
    progr = (delta / float(pwm_func_period)) % 1
    if progr < 0.5:
        freq = pwm_func_freqmin
        bzr_passive_1.freq(int(freq))
    else:
        freq = pwm_func_freqmax
        bzr_passive_1.freq(int(freq))


def cmd_worker():
    global cmd_queue
    global pwm_func_start
    global pwm_func_period
    global pwm_func_freqmin
    global pwm_func_freqmax

    
    if len(cmd_queue) == 0:
        return
    
    nowticks = time.ticks_ms()
    
    try:
            
        if cmd_queue[0][0] >= nowticks:
            return
    
        cmdarr = cmd_queue.pop(0)
        tcmd = cmdarr[0]
        icmd = cmdarr[1]
        cmd = cmdarr[2]
        print("[EXEC] Execute: %d @ %d (cmd:%d) -> %s" % (icmd, nowticks, tcmd, str(cmd)))
                
        if icmd == CMD_RESET:
            key.value(1)
            bzr_active_1.value(0)
            bzr_active_2.value(0)
            bzr_active_3.value(0)
            bzr_passive_1.duty_u16(0)
            led_r.value(0)
            led_g.value(0)
            led_b.value(0)
            buzz_fn_worker.deinit()
            cmd_queue = []
        elif icmd == CMD_KEY_ON:
            key.value(0) # open drain - inverted
        elif icmd == CMD_KEY_OFF:
            key.value(1) # open drain - inverted
        elif icmd == CMD_BZR1_ON:
            bzr_active_1.value(1)
        elif icmd == CMD_BZR1_OFF:
            bzr_active_1.value(0)
        elif icmd == CMD_BZR2_ON:
            bzr_active_2.value(1)
        elif icmd == CMD_BZR2_OFF:
            bzr_active_2.value(0)
        elif icmd == CMD_BZR3_ON:
            bzr_active_3.value(1)
        elif icmd == CMD_BZR3_OFF:
            bzr_active_3.value(0)
        elif icmd == CMD_PWM_ON:
            bzr_passive_1.freq(int(cmd['frequency']))
            bzr_passive_1.duty_u16(bzr_passive_duty)
            bzr_passive_1.freq(int(cmd['frequency']))
        elif icmd == CMD_PWM_OFF:
            bzr_passive_1.duty_u16(0)
        elif icmd == CMD_PWM_FUNC_ON:
            pwm_func_start = time.time_ns()
            pwm_func_period = int(cmd['period']) * 1_000_000
            pwm_func_freqmin = int(cmd['frequency_min'])
            pwm_func_freqmax = int(cmd['frequency_max'])
            if cmd['func'] == 'SINUS':
                pwm_func_sinus(None)
                buzz_fn_worker.init(period=16, mode=Timer.PERIODIC, callback=pwm_func_sinus)
                bzr_passive_1.duty_u16(bzr_passive_duty)
            if cmd['func'] == 'TRIANGLE':
                pwm_func_triangle(None)
                buzz_fn_worker.init(period=16, mode=Timer.PERIODIC, callback=pwm_func_triangle)
                bzr_passive_1.duty_u16(bzr_passive_duty)
            if cmd['func'] == 'SAWTOOTH':
                pwm_func_sawtooth(None)
                buzz_fn_worker.init(period=16, mode=Timer.PERIODIC, callback=pwm_func_sawtooth)
                bzr_passive_1.duty_u16(bzr_passive_duty) # 50% on, 50% off
            if cmd['func'] == 'SQUARE':
                pwm_func_square(None)
                buzz_fn_worker.init(period=16, mode=Timer.PERIODIC, callback=pwm_func_square)
                bzr_passive_1.duty_u16(bzr_passive_duty) # 50% on, 50% off
            pass
        elif icmd == CMD_PWM_FUNC_OFF:
            buzz_fn_worker.deinit()
            bzr_passive_1.duty_u16(0)
        elif icmd == CMD_PWM_NOTE_ON:
            bzr_passive_1.freq(cmd['frequency'])
            bzr_passive_1.duty_u16(bzr_passive_duty) # 50% on, 50% off
        elif icmd == CMD_PWM_NOTE:
            bzr_passive_1.freq(cmd['frequency'])
        elif icmd == CMD_PWM_NOTE_OFF:
            bzr_passive_1.duty_u16(0)
        else:
            raise ValueError("unknown icmd: " + str(cmd[1]))
            
            
    except Exception as e:
        print('Error while executing: ' + str(e))
        sys.print_exception(e)
        led_blink(led_r, 2, 0.1, 0.1)
        time.sleep_ms(5)
        return
    
        
    
    pass



################  ################  ################  ################  ################  ################  ##################

wifi_connect()

################  ################  ################  ################  ################  ################  ##################

delta_timer_wifi = 3000
last_timer_wifi  = -delta_timer_wifi
enabl_timer_wifi = 1

delta_timer_cmdquery = 3000
last_timer_cmdquery  = -delta_timer_cmdquery
enabl_timer_cmdquery = 1

delta_timer_cmdworker = 32
last_timer_cmdworker  = -delta_timer_cmdworker
enabl_timer_cmdworker = 1

t0 = time.ticks_ms()

################  ################  ################  ################  ################  ################  ##################

while True:
    
    try:
        
        ticks = time.ticks_ms()
        
        if enabl_timer_wifi == 1 and (ticks - last_timer_wifi) > delta_timer_wifi:
            wifi_worker()
            last_timer_wifi = ticks = time.ticks_ms()
            
        if enabl_timer_cmdquery == 1 and last_wifi_status == network.STAT_GOT_IP and (((ticks - last_timer_cmdquery) > delta_timer_cmdquery) or len(cmd_queue) == 0):
            cmd_query_worker()
            last_timer_cmdquery = ticks = time.ticks_ms()
            
        if enabl_timer_cmdworker == 1 and ((ticks - last_timer_cmdworker) > delta_timer_cmdworker or ((len(cmd_queue) > 0) and (cmd_queue[0][0] < (nowticks - delta_timer_cmdworker*2) )) ):
            cmd_worker()
            last_timer_cmdworker = ticks = time.ticks_ms()

    except Exception as e:
        print('Error in main loop: ' + str(e))
        sys.print_exception(e)
        led_blink(led_r, 8, 0.1, 0.1)
        time.sleep_ms(5)
        continue 



