from pathlib import Path
import base64, mimetypes, subprocess, asyncio
from jinja2 import Environment,FileSystemLoader
from playwright.async_api import async_playwright
root=Path('/Users/flanchan/Developer/Projects/GithubProjects')
tpl=root/'astrbot-plugin-dev/data/plugins/astrbot_plugin_dna/src/templates'
assets=root/'dna-resource'
out=Path('/tmp/alonabot-dna-compare');out.mkdir(exist_ok=True)
env=Environment(loader=FileSystemLoader(str(tpl)),autoescape=True)
def uri(p):
 p=assets/p
 return 'data:'+ (mimetypes.guess_type(p.name)[0] or 'image/png')+';base64,'+base64.b64encode(p.read_bytes()).decode()
font=uri('fonts/dna_fonts.ttf')
icons=['mh_role','mh_weapon','mh_mzx']
entries=[{'icon':uri('textures/mh/'+i+'.png'),'type_name':n,'instances':[{'name':s,'subscribed':j==0} for j,s in enumerate(names)]} for i,n,names in zip(icons,['角色密函','武器密函','魔之楔密函'],[['辉夜试炼','世界树之梦','无尽回廊'],['星轨漫游','武器突破','秘境挑战'],['守护者之路','裂隙挑战','探索行动']])]
header={'avatar':uri('images/role_avatar/1101.png'),'avatar_frame':uri('textures/common/avatar_frame.png'),'name':'测试玩家','uid':'******1234','level':65,'level_background':uri('textures/common/avatar_title_level.png'),'stats':[{'value':'42','label':'角色数量'},{'value':'26','label':'武器数量'}],'stats_background':uri('textures/common/avatar_title_base_info.png')}
tasks=[
 ('mh', 'cards/mh_card.html.j2',dict(width=1700,height=900,font=font,background=uri('textures/mh/bg1.jpg'),bar=uri('textures/mh/bar.png'),cards=entries,card_background=uri('textures/mh/card.png'),footer=uri('textures/common/footer.png'),refresh_text='23分钟后刷新',refresh_background=uri('textures/mh/refresh_time.png'),title=uri('textures/mh/title.png')),1700,900),
 ('mh_simple','cards/mh_simple.html.j2',dict(width=1040,font=font,entries=entries,refresh_text='09:00 - 10:00    23分钟后刷新'),1040,646),
 ('role_info','cards/role_info.html.j2',dict(width=1200,height=1160,font=font,background=uri('textures/common/bg.jpg'),title_mask=uri('textures/role/title_mask.png'),title_background=uri('textures/role/title_bg.jpg'),header=header,header_background=uri('textures/common/avatar_title_bg.png'),info_bar=uri('textures/role/info_bar.png'),achievements=[{'value':'42','label':'已解锁角色'},{'value':'26','label':'已收集武器'},{'value':'180','label':'探索进度'},{'value':'35','label':'任务完成'}],div_background=uri('textures/role/div_bg.png'),sections=[{'title':'已解锁角色','background':uri('textures/role/div_bg.png'),'items':[{'name':n,'image':uri('images/role_avatar/'+id+'.png'),'element':'','grade':'','level':lv,'unlocked':True,'type':'role'} for n,id,lv in [('角色一','1101',80),('角色二','1102',70),('角色三','1103',65),('角色四','1201',60),('角色五','1501',55)]]}],item_mask=uri('textures/role/item_mask.png'),item_foreground=uri('textures/role/item_fg.png'),footer_image=uri('textures/common/footer.png')),1200,1160)
]
async def main():
 async with async_playwright() as p:
  browser=await p.chromium.launch(executable_path='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless=True)
  for name,template,data,w,h in tasks:
   html=env.get_template(template).render(**data)
   (out/(name+'.html')).write_text(html)
   page=await browser.new_page(viewport={'width':w,'height':h},device_scale_factor=1)
   await page.set_content(html,wait_until='load',timeout=60000)
   await page.screenshot(path=str(out/(name+'_chrome.png')),full_page=False)
   await page.close()
   result=subprocess.run(['/tmp/alonabot-ogre-cli','--render',str(out/(name+'.html')),'--width',str(w),'--height',str(h),'--format','png','--output',str(out/(name+'_ogre.png'))],capture_output=True,text=True,timeout=120)
   print(name,'chrome=',(out/(name+'_chrome.png')).stat().st_size,'ogre_exit=',result.returncode,'ogre_bytes=',(out/(name+'_ogre.png')).stat().st_size if (out/(name+'_ogre.png')).exists() else None,'err=',result.stderr[:500],flush=True)
  await browser.close()
asyncio.run(main())
