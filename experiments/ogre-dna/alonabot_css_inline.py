from pathlib import Path
from lxml import html
import tinycss2,cssselect2
src=Path('/tmp/alonabot-dna-compare')
for name in ('mh_simple','mh','role_info'):
 root=html.fromstring((src/(name+'.html')).read_text())
 matcher=cssselect2.Matcher()
 for style in root.xpath('//style'):
  stylesheet=style.text or ''
  for rule in tinycss2.parse_stylesheet(stylesheet,skip_whitespace=True,skip_comments=True):
   if rule.type!='qualified-rule':continue
   declarations=[d for d in tinycss2.parse_declaration_list(rule.content,skip_whitespace=True,skip_comments=True) if d.type=='declaration']
   try:
    selectors=cssselect2.compile_selector_list(rule.prelude)
   except cssselect2.SelectorError:continue
   for selector in selectors:
    if selector.pseudo_element is None:matcher.add_selector(selector,declarations)
  style.getparent().remove(style)
 count=0
 for wrapper in cssselect2.ElementWrapper.from_html_root(root).iter_subtree():
  el=wrapper.etree_element
  original=el.attrib.get('style','')
  rules={}
  for specificity,order,pseudo,declarations in matcher.match(wrapper):
   for d in declarations:
    priority=(1 if d.important else 0, 0, *specificity, order)
    value=tinycss2.serialize(d.value).strip()
    if d.name not in rules or priority>=rules[d.name][0]:rules[d.name]=(priority,value)
  for d in tinycss2.parse_declaration_list(original,skip_whitespace=True,skip_comments=True):
   if d.type!='declaration':continue
   priority=(1 if d.important else 0,1,0,0,0,0)
   if d.name not in rules or priority>=rules[d.name][0]:rules[d.name]=(priority,tinycss2.serialize(d.value).strip())
  if rules:
   el.set('style',';'.join(f'{k}:{v[1]}' for k,v in rules.items()))
   count+=1
  elif original:
   el.attrib.pop('style')
 target=src/(name+'_inlined.html')
 target.write_bytes(html.tostring(root,encoding='utf-8',method='html',doctype='<!doctype html>'))
 print(name,'elements',count,'size_mb',round(target.stat().st_size/1e6,1),flush=True)
