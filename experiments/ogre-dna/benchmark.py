import asyncio
import json
import re
import socket
import subprocess
import tempfile
import threading
import time
from pathlib import Path
from statistics import median
from urllib.request import urlopen
from playwright.async_api import async_playwright

ROOT = Path("/Users/flanchan/alonabot/experiments/ogre-dna")
IN = Path("/tmp/alonabot-dna-compare")
FONT = "/Users/flanchan/Developer/Projects/GithubProjects/dna-resource/fonts/dna_fonts.ttf"
CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
COUNT = 10

def process_table():
    table = {}
    rows = subprocess.check_output(
        ["ps", "-axo", "pid=,ppid=,rss=,time=,command="], text=True
    ).splitlines()
    for row in rows:
        parts = row.strip().split(None, 4)
        if len(parts) != 5:
            continue
        try:
            pid, ppid, rss = map(int, parts[:3])
        except ValueError:
            continue
        value = parts[3].split(":")
        seconds = float(value[-1])
        for i, amount in enumerate(reversed(value[:-1]), start=1):
            seconds += int(amount) * (60 ** i)
        table[pid] = dict(ppid=ppid, rss=rss, cpu=seconds, command=parts[4])
    return table

def proc_tree_stats(root_pid, table=None):
    table = table or process_table()
    selected = {root_pid}
    while True:
        prev = len(selected)
        selected.update(pid for pid, row in table.items() if row["ppid"] in selected)
        if prev == len(selected):
            break
    return {
        "processes": len(selected),
        "rss_mb": round(sum(row["rss"] for pid,row in table.items() if pid in selected)/1024,1),
        "cpu_s": round(sum(row["cpu"] for pid,row in table.items() if pid in selected),3)
    }

class MemoryMonitor:
    def __init__(self, root):
        self.root = root
        self.stop = threading.Event()
        self.thread = None
        self.peak = 0
        self.last = 0
    def loop(self):
        while not self.stop.is_set():
            try:
                value = proc_tree_stats(self.root)["rss_mb"]
                self.peak = max(value, self.peak)
                self.last = value
            except Exception:
                pass
            self.stop.wait(0.14)
    def start(self):
        self.thread = threading.Thread(target=self.loop, daemon=True)
        self.thread.start()
    def end(self):
        self.stop.set()
        self.thread.join()
        return {"peak_rss_mb":self.peak, "final_rss_mb":self.last}

def summarize(times):
    return {
        "first_ms":round(times[0]*1000,1),
        "warm_median_ms":round(median(times[1:])*1000,1),
        "warm_min_ms":round(min(times[1:])*1000,1),
        "warm_max_ms":round(max(times[1:])*1000,1),
        "all_ms":[round(x*1000,1) for x in times],
    }

async def main():
    results={}
    with socket.socket() as sock:
        sock.bind(("127.0.0.1",0))
        port=sock.getsockname()[1]
    args = ["/tmp/alonabot-ogre-bench-server", f"127.0.0.1:{port}",
            str(IN / "mh_ogre_adapted.html"), FONT, str(ROOT / "ogre_bench.png")]
    start = time.perf_counter()
    go = subprocess.Popen(args,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    try:
        ready=go.stdout.readline().strip()
        if ready!="READY":
            raise RuntimeError(f"Go render server failed to start: {ready}")
        start_ms=(time.perf_counter()-start)*1000
        go_monitor=MemoryMonitor(go.pid)
        go_monitor.start()
        base=proc_tree_stats(go.pid)
        times=[]
        for _ in range(COUNT):
            begin=time.perf_counter()
            with urlopen(f"http://127.0.0.1:{port}/render",timeout=50) as response:
                data=response.read()
                if not data.startswith(b"\x89PNG"):raise RuntimeError("Go did not return PNG")
            times.append(time.perf_counter()-begin)
        end=proc_tree_stats(go.pid)
        results["ogre"] = {"startup_ms":round(start_ms,1),
                           **summarize(times),
                           **go_monitor.end(),
                           "steady_rss_mb":end["rss_mb"],
                           "cpu_s_all_10":round(end["cpu_s"]-base["cpu_s"],3),
                           "processes":end["processes"],
                           "png_bytes":len(data)}
        print("OGRE",json.dumps(results["ogre"],ensure_ascii=False),flush=True)
        footprint=subprocess.run(["/usr/bin/footprint","--noCategories","--pid",str(go.pid)],capture_output=True,text=True,timeout=30)
        print("GO_PHYSICAL",footprint.returncode,(footprint.stdout+footprint.stderr)[-2800:],flush=True)
    finally:
        go.terminate()
        try: go.wait(timeout=4)
        except subprocess.TimeoutExpired:go.kill()

    profile = tempfile.mkdtemp(prefix="alonabot-chrome-bench-")
    async with async_playwright() as play:
        startup=time.perf_counter()
        ctx=await play.chromium.launch_persistent_context(
             user_data_dir=profile, executable_path=CHROME, headless=True,
             viewport={"width":1700,"height":900},device_scale_factor=1,
             args=["--disable-background-networking","--no-first-run"])
        browser_start_ms=(time.perf_counter()-startup)*1000
        page=ctx.pages[0] if ctx.pages else await ctx.new_page()
        chrome_pid=None
        for attempt in range(30):
            procs=process_table()
            candidates=[
                pid for pid, row in procs.items()
                if profile in row["command"]
                and "Google Chrome" in row["command"]
                and "--type=" not in row["command"]
            ]
            if candidates:chrome_pid=candidates[0];break
            await asyncio.sleep(0.1)
        if chrome_pid is None: raise RuntimeError("Chrome root process not found")
        chrome_monitor=MemoryMonitor(chrome_pid)
        chrome_monitor.start()
        base=proc_tree_stats(chrome_pid)
        times=[]
        target=(IN/"mh.html").as_uri()
        for _ in range(COUNT):
            begin=time.perf_counter()
            await page.goto(target, wait_until="load",timeout=50000)
            await page.evaluate("document.fonts.ready")
            screenshot=await page.screenshot(type="png",animations="disabled")
            times.append(time.perf_counter()-begin)
        end=proc_tree_stats(chrome_pid)
        results["chromium"]={"startup_ms":round(browser_start_ms,1),
                             **summarize(times),
                             **chrome_monitor.end(),
                             "steady_rss_mb":end["rss_mb"],
                             "cpu_s_all_10":round(end["cpu_s"]-base["cpu_s"],3),
                             "processes":end["processes"],
                             "png_bytes":len(screenshot)}
        print("CHROME",json.dumps(results["chromium"],ensure_ascii=False),flush=True)
        table=process_table(); descendants={chrome_pid}
        while True:
            n=len(descendants)
            descendants.update(pid for pid,row in table.items() if row["ppid"] in descendants)
            if n==len(descendants):break
        footprint=subprocess.run(["/usr/bin/footprint","--noCategories",*[v for pid in sorted(descendants) for v in ("--pid",str(pid))]],capture_output=True,text=True,timeout=30)
        print("CHROME_PHYSICAL",footprint.returncode,(footprint.stdout+footprint.stderr)[-3500:],flush=True)
        await ctx.close()
    (ROOT/"benchmark.json").write_text(json.dumps(results,ensure_ascii=False,indent=2),encoding="utf-8")
    print("RESULT_FILE",ROOT/"benchmark.json",flush=True)

asyncio.run(main())
