import { useEffect, useRef, useContext, type JSX } from 'react';
import { bindPluginPage } from '../bridge/pluginPage';
import { SearchContext } from '../shell/search';
import { useUISettings } from '../shell/settings';

export function PluginPageFrame(props: { id: string; path: string; title: string }): JSX.Element {
  const { id, path, title } = props;
  const { locale } = useUISettings();
  const { open } = useContext(SearchContext);
  const frameRef = useRef<HTMLIFrameElement>(null);
  const bridgeRef = useRef<ReturnType<typeof bindPluginPage> | null>(null);
  const localeRef = useRef(locale);
  localeRef.current = locale;
  useEffect(() => {
    const frame = frameRef.current!;
    const bind = () => {
      bridgeRef.current?.dispose();
      bridgeRef.current = bindPluginPage(frame, { id, locale: localeRef.current, openSearch: open });
    };
    bind();
    frame.addEventListener('load', bind);
    return () => { frame.removeEventListener('load', bind); bridgeRef.current?.dispose(); bridgeRef.current = null; };
  }, [id, path, open]);
  useEffect(() => { bridgeRef.current?.setLocale(locale); }, [locale]);
  return <iframe ref={frameRef} className="plugin-page-frame" src={`/plugins/${encodeURIComponent(id)}/${path.replace(/^\//, '')}`} title={title} sandbox="allow-scripts allow-forms allow-popups" />;
}
