from lxml import html
from pathlib import Path
from base64 import b64encode
src=Path('/tmp/alonabot-dna-compare')
assets=Path('/Users/flanchan/Developer/Projects/GithubProjects/dna-resource/textures/mh')
bar='data:image/png;base64,'+b64encode((assets/'bar.png').read_bytes()).decode('ascii')
refresh_icon='data:image/png;base64,'+b64encode((assets/'refresh_time.png').read_bytes()).decode('ascii')
root=html.fromstring((src/'mh_inlined.html').read_bytes())
for parent in root.xpath('//*[contains(concat(" ",normalize-space(@class)," ")," mh-card__instance-wrap ")]'):
 parent.set('style',parent.get('style','')+';display:block;position:absolute;width:374px;height:295px')
 for i,li in enumerate(parent.xpath('./li')):
  li.set('style',li.get('style','')+f';position:absolute;left:0;top:{i*80}px')
  text=li.text
  li.text=None
  img=html.Element('img');img.set('src',bar);img.set('alt','');img.set('style','position:absolute;left:0;top:0;width:374px;height:55px;object-fit:fill;z-index:0')
  li.append(img)
  img.tail=text

for notice in root.xpath('//*[contains(concat(" ",normalize-space(@class)," ")," mh-card__refresh ")]'):
 img=html.Element('img')
 img.set('src',refresh_icon)
 img.set('alt','')
 img.set('style','position:absolute;left:0;top:0;width:317px;height:76px;object-fit:fill;z-index:0')
 img.tail=notice.text
 notice.text=None
 notice.insert(0,img)

out=src/'mh_ogre_adapted.html'
out.write_bytes(html.tostring(root,encoding='utf-8',method='html',doctype='<!doctype html>'))
