# Python script to fix parameter modal form input sync, textarea rendering, and cancel button styling

with open('frontend/src/App.jsx', 'r', encoding='utf-8') as f:
    app_code = f.read()

# 1. Update getPrimaryKeyKey for parameters to be 'key_name' or 'id'
old_pk_key = """    if (tab === 'parameters') return 'id';"""
new_pk_key = """    if (tab === 'parameters') return 'key_name';"""
app_code = app_code.replace(old_pk_key, new_pk_key)

# 2. Update getFieldValue helper to check both masterForm and paramForm
old_get_val_helper = """  const getFieldValue = (form, key) => {
    if (!form) return '';
    if (form[key] !== undefined && form[key] !== null) return form[key];
    const match = Object.keys(form).find(m => m.toLowerCase() === key.toLowerCase());
    if (match && form[match] !== undefined && form[match] !== null) return form[match];
    return '';
  };"""

new_get_val_helper = """  const getFieldValue = (form, key) => {
    if (form && form[key] !== undefined && form[key] !== null) return form[key];
    if (paramForm && paramForm[key] !== undefined && paramForm[key] !== null) return paramForm[key];
    if (form) {
      const match = Object.keys(form).find(m => m.toLowerCase() === key.toLowerCase());
      if (match && form[match] !== undefined && form[match] !== null) return form[match];
    }
    if (paramForm) {
      const match = Object.keys(paramForm).find(m => m.toLowerCase() === key.toLowerCase());
      if (match && paramForm[match] !== undefined && paramForm[match] !== null) return paramForm[match];
    }
    return '';
  };"""

app_code = app_code.replace(old_get_val_helper, new_get_val_helper)

# 3. Add textarea input rendering and update input onChange in modal form
old_input_render = """                                  <input 
                                    type={f.type === 'idr' ? 'text' : (f.type || 'text')}
                                    disabled={isDisabled}
                                    required={f.type !== 'checkbox' && (f.k !== 'password' || !isEditMode) && f.k !== 'member_no'}
                                    checked={f.type === 'checkbox' ? Boolean(val) : undefined}
                                    value={f.type === 'idr' ? formatRupiahInput(val) : (f.type !== 'checkbox' ? (val ?? '') : undefined)}
                                    onChange={e => setMasterForm({
                                      ...masterForm, 
                                      [f.k]: f.type === 'checkbox' ? e.target.checked : (f.type === 'idr' ? parseRupiahInput(e.target.value) : (f.type === 'number' ? parseInt(e.target.value) || 0 : e.target.value))
                                    })}"""

new_input_render = """                                  f.type === 'textarea' ? (
                                    <textarea 
                                      disabled={isDisabled}
                                      required
                                      value={val ?? ''}
                                      onChange={e => {
                                        const newVal = e.target.value;
                                        setMasterForm(prev => ({ ...prev, [f.k]: newVal }));
                                        setParamForm(prev => ({ ...prev, [f.k]: newVal }));
                                      }}
                                      style={{ width: '100%', padding: '10px', borderRadius: '6px', border: '1px solid #cbd5e1', minHeight: '80px', fontFamily: 'inherit' }}
                                    />
                                  ) : (
                                    <input 
                                      type={f.type === 'idr' ? 'text' : (f.type || 'text')}
                                      disabled={isDisabled}
                                      required={f.type !== 'checkbox' && (f.k !== 'password' || !isEditMode) && f.k !== 'member_no'}
                                      checked={f.type === 'checkbox' ? Boolean(val) : undefined}
                                      value={f.type === 'idr' ? formatRupiahInput(val) : (f.type !== 'checkbox' ? (val ?? '') : undefined)}
                                      onChange={e => {
                                        const newVal = f.type === 'checkbox' ? e.target.checked : (f.type === 'idr' ? parseRupiahInput(e.target.value) : (f.type === 'number' ? parseInt(e.target.value) || 0 : e.target.value));
                                        setMasterForm(prev => ({ ...prev, [f.k]: newVal }));
                                        setParamForm(prev => ({ ...prev, [f.k]: newVal }));
                                      }}"""

app_code = app_code.replace(old_input_render, new_input_render)

# Close extra parenthesis if textarea ternary added
old_closing_input = """                                    style={f.type !== 'checkbox' ? { 
                                      width: '100%', 
                                      padding: '10px', 
                                      borderRadius: '6px', 
                                      border: '1px solid #cbd5e1',
                                      backgroundColor: isDisabled ? '#f1f5f9' : '#ffffff',
                                      color: isDisabled ? '#64748b' : '#0f172a',
                                      fontWeight: isDisabled ? 600 : 400,
                                      cursor: isDisabled ? 'not-allowed' : 'text'
                                    } : { width: '20px', height: '20px' }}
                                  />
                                )}"""

new_closing_input = """                                    style={f.type !== 'checkbox' ? { 
                                      width: '100%', 
                                      padding: '10px', 
                                      borderRadius: '6px', 
                                      border: '1px solid #cbd5e1',
                                      backgroundColor: isDisabled ? '#f1f5f9' : '#ffffff',
                                      color: isDisabled ? '#64748b' : '#0f172a',
                                      fontWeight: isDisabled ? 600 : 400,
                                      cursor: isDisabled ? 'not-allowed' : 'text'
                                    } : { width: '20px', height: '20px' }}
                                  />
                                  )}"""

app_code = app_code.replace(old_closing_input, new_closing_input)

# Update Cancel button styling in modal form to strictly follow BUTTON_CLOSED_BG and font white
old_modal_cancel_btn = """                        <button
                          type="button"
                          onClick={() => {
                            setIsMasterModalOpen(false);
                            setIsEditMasterMode(false);
                            setMasterForm({});
                            setParamForm({ id: 0, key_name: '', key_value: '', description: '' });
                          }}
                          style={{ padding: '10px 22px', background: 'var(--button-closed-bg, #f59e0b)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                        >
                          Cancel
                        </button>"""

new_modal_cancel_btn = """                        <button
                          type="button"
                          onClick={() => {
                            setIsMasterModalOpen(false);
                            setIsEditMasterMode(false);
                            setMasterForm({});
                            setParamForm({ id: 0, key_name: '', key_value: '', description: '' });
                          }}
                          style={{ padding: '10px 22px', background: 'var(--button-closed-bg, #f59e0b)', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', width: 'auto' }}
                        >
                          Cancel
                        </button>"""

app_code = app_code.replace(old_modal_cancel_btn, new_modal_cancel_btn)

with open('frontend/src/App.jsx', 'w', encoding='utf-8') as f:
    f.write(app_code)

print("Successfully updated frontend/src/App.jsx with synchronized parameter modal form & textarea!")
