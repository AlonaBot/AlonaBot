import { useId, useMemo, useState } from 'react';
import type { ChangeEvent, FormEvent } from 'react';
import { fieldsFromSchema } from '../schema/model';
import type { Field } from '../schema/model';
import { buildPatch, isSecretSet, MissingSecretsError } from '../schema/patch';
import { fieldText } from '../i18n/catalog';
import { useUI } from '../i18n/UI';

function valuePath(path: string[]): string {
  return '/' + path.map(part => part.replaceAll('~', '~0').replaceAll('/', '~1')).join('/');
}

type FieldProps = {
  field: Field;
  value: unknown;
  original: unknown;
  path: string[];
  messages: Record<string, unknown>;
  missing: string[];
  onChange: (value: unknown) => void;
};

function FieldControl({ field, value, original, path, messages, missing, onChange }: FieldProps) {
  const ui = useUI();
  const id = useId();
  const [editingSecret, setEditingSecret] = useState(false);
  const text = fieldText(messages, field);
  const missingValue = missing.includes(valuePath(path));
  const attributes = { id, 'aria-describedby': text.hint ? `${id}-hint` : undefined, 'aria-invalid': missingValue || undefined };
  let control;

  if (value === null) {
    control = <div className="default-pending" role="status">
      <p>{ui.defaultPending}</p>
      <button type="button" onClick={() => onChange(original)}>{ui.undoReset}</button>
    </div>;
  } else if (field.secret) {
    const set = isSecretSet(value) || (typeof value === 'string' && value.length > 0);
    control = <div className="secret-control">
      <span role="status">{set ? ui.secretSet : ui.secretUnset}</span>
      {!editingSecret && <button type="button" onClick={() => setEditingSecret(true)}>{ui.modify}</button>}
      {editingSecret && <input {...attributes} type="password" autoComplete="new-password"
        required={field.required} value={typeof value === 'string' ? value : ''}
        onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.currentTarget.value)} />}
      <button type="button" onClick={() => { setEditingSecret(false); onChange(null); }}>{ui.clear}</button>
    </div>;
  } else {
    switch (field.type) {
      case 'object':
        control = <Fields fields={field.children!} value={value as Record<string, unknown>}
          original={original as Record<string, unknown>} path={path} messages={messages} missing={missing} onChange={onChange} />;
        break;
      case 'bool':
        control = <input {...attributes} type="checkbox" role="switch" checked={value === true}
          onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.currentTarget.checked)} />;
        break;
      case 'int':
      case 'float':
        control = <div className="number-control">
          <input {...attributes} type={field.slider ? 'range' : 'number'} required={value !== undefined} value={value === undefined ? '' : value as number | string}
            min={field.slider?.min} max={field.slider?.max} step={field.slider?.step ?? (field.type === 'int' ? 1 : 'any')}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.currentTarget.value === '' ? '' : event.currentTarget.valueAsNumber)} />
          {field.slider && <output htmlFor={id}>{String(value)}</output>}
        </div>;
        break;
      case 'string':
      case 'text':
        if (field.editorMode || field.type === 'text') {
          control = <div>
            {field.editorMode && <span className="editor-language">{ui.codeLanguage}: {field.editorLanguage}</span>}
            <textarea {...attributes} className={field.editorMode ? 'code-editor' : undefined}
              style={field.editorMode ? { fontFamily: 'monospace' } : undefined}
              required={field.required} value={typeof value === 'string' ? value : ''} spellCheck={!field.editorMode}
              onChange={(event: ChangeEvent<HTMLTextAreaElement>) => onChange(event.currentTarget.value)} />
          </div>;
        } else if (field.options) {
          control = <select {...attributes} required={field.required} value={typeof value === 'string' ? value : ''}
            onChange={(event: ChangeEvent<HTMLSelectElement>) => onChange(event.currentTarget.value)}>
            {field.options.map((option, index) => <option key={option} value={option}>{text.labels?.[index] ?? option}</option>)}
          </select>;
        } else {
          control = <input {...attributes} type="text" required={field.required} value={typeof value === 'string' ? value : ''}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.currentTarget.value)} />;
        }
        break;
      default:
        throw new Error(`unsupported form control: ${field.type}`);
    }
  }

  const title = <>{text.description}{field.required && <span className="required-mark"> · {ui.required}</span>}</>;
  const body = <>
    {field.type === 'object' && <legend>{title}</legend>}
    <div className="field-heading">
      {field.type !== 'object' && <label htmlFor={id}>{title}</label>}
      <button type="button" onClick={() => { setEditingSecret(false); onChange(null); }}>{ui.restoreDefault}</button>
    </div>
    {text.hint && <p id={`${id}-hint`} className={field.obviousHint ? 'field-hint obvious-hint' : 'field-hint'}
      style={field.obviousHint ? { fontWeight: 700 } : undefined}>{text.hint}</p>}
    {control}
    {missingValue && <p role="alert">{ui.missingSecrets} {valuePath(path)}</p>}
  </>;
  return field.type === 'object' ? <fieldset className="config-group">{body}</fieldset> : <div className="config-field">{body}</div>;
}

function Fields({ fields, value, original, path, messages, missing, onChange }: {
  fields: Field[];
  value: Record<string, unknown>;
  original: Record<string, unknown>;
  path: string[];
  messages: Record<string, unknown>;
  missing: string[];
  onChange: (value: unknown) => void;
}) {
  return <>{fields.filter(field => !field.invisible && (!field.condition || Object.entries(field.condition)
    .every(([key, expected]) => JSON.stringify(value[key]) === JSON.stringify(expected))))
    .map(field => <FieldControl key={field.key} field={field} value={value[field.key]} original={original?.[field.key]}
      path={[...path, field.key]} messages={messages} missing={missing}
      onChange={next => onChange({ ...value, [field.key]: next })} />)}</>;
}

type ConfigFormProps = {
  schema: unknown;
  value: unknown;
  messages: Record<string, unknown>;
  onSubmit: (patch: Record<string, unknown>) => void;
};

/** ConfigForm 渲染配置字段并在提交时生成补丁。 */
export function ConfigForm(props: ConfigFormProps) {
  const ui = useUI();
  const fields = useMemo(() => fieldsFromSchema(props.schema), [props.schema]);
  const [edited, setEdited] = useState(() => structuredClone(props.value) as Record<string, unknown>);
  const [error, setError] = useState<Error | null>(null);
  const missing = error instanceof MissingSecretsError ? error.paths : [];

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!event.currentTarget.reportValidity()) return;
    let patch: Record<string, unknown>;
    try {
      patch = buildPatch(props.value, edited, fields);
    } catch (caught) {
      if (!(caught instanceof Error)) throw caught;
      setError(caught);
      return;
    }
    setError(null);
    props.onSubmit(patch);
  }

  return <form onSubmit={submit}>
    <Fields fields={fields} value={edited} original={props.value as Record<string, unknown>}
      path={[]} messages={props.messages} missing={missing} onChange={next => { setEdited(next as Record<string, unknown>); setError(null); }} />
    {error && <div role="alert" className="form-error">
      <p>{error instanceof MissingSecretsError ? ui.missingSecrets : ui.patchError}</p>
      <pre>{error.message}</pre>
    </div>}
    <button type="submit" className="save-button">{ui.save}</button>
  </form>;
}
