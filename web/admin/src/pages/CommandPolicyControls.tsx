import { Edit2, RotateCcw } from 'lucide-react';
import type { CommandOverride, CommandScope } from '../api/admin';
import { useUI } from '../i18n/UI';
import { commandOverride, permissionPreset } from './permissionState';

type Change = Partial<Pick<CommandOverride, 'enabled' | 'permissions'>>;

export function CommandPolicyControls({ overrides, scope, enabled, permissions, label, busy, onChange, onCustom }: {
  overrides: CommandOverride[]; scope: CommandScope; enabled: boolean; permissions?: string[] | null;
  label: string; busy: boolean; onChange: (change: Change) => void; onCustom: (permissions: string[] | null) => void;
}) {
  const ui = useUI();
  const override = commandOverride(overrides, scope, {});
  const preset = permissionPreset(override.permissions);
  const labels: Record<string, string> = { everyone: ui.permEveryone, chat_admin: ui.permGroupAdmin, admin: ui.permAdminOnly, custom: permissions?.join(', ') || ui.customPermissions };
  return <div className="command-policy-controls">
    <select aria-label={`${ui.permissions} ${label}`} disabled={busy} value={preset} onChange={event => {
      const value = event.target.value;
      if (value === 'custom') onCustom(override.permissions);
      else onChange({ permissions: value === 'inherit' ? null : value === 'everyone' ? [] : [value] });
    }}>
      <option value="inherit">{ui.commandInherit}{permissions !== undefined ? ` (${labels[permissionPreset(permissions ?? [])]})` : ''}</option>
      <option value="everyone">{ui.permEveryone}</option><option value="chat_admin">{ui.permGroupAdmin}</option><option value="admin">{ui.permAdminOnly}</option>
      <option value="custom">{preset === 'custom' ? `${ui.customPermissions}: ${override.permissions?.join(', ')}` : ui.customPermissions}</option>
    </select>
    {preset === 'custom' && <button type="button" className="detail-icon-button" aria-label={`${ui.customPermissions} ${label}`} disabled={busy} onClick={() => onCustom(override.permissions)}><Edit2 size={14} /></button>}
    <div className="command-enabled-control">
      <label className="plugin-switch-label"><input type="checkbox" role="switch" aria-label={`${ui.commandEnabled} ${label}`} title={override.enabled === null ? ui.commandInherit : undefined} checked={enabled} disabled={busy} onChange={event => onChange({ enabled: event.target.checked })} /><span className="plugin-switch-slider" /></label>
      <button type="button" className="command-inherit-button" disabled={busy || override.enabled === null} title={ui.resetEnabled} aria-label={`${ui.resetEnabled} ${label}`} onClick={() => onChange({ enabled: null })}>{override.enabled === null ? ui.commandInherit : <RotateCcw size={12} />}</button>
    </div>
  </div>;
}
