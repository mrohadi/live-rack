/* global React, LR_DATA */
const Dx = window.LR_DATA;

// ============ SCANNER (mobile + desktop split) ============
function Scanner({ accent }) {
  const [scans, setScans] = React.useState([
    { sku: 'LR-7821', name: 'Merino Crew Sweater', zone: 'A2', t: '14:24:08', ok: true },
    { sku: 'LR-3318', name: 'Wireless Earbuds',    zone: 'A4', t: '14:23:55', ok: true },
    { sku: 'LR-8870', name: 'Smart Bulb 4-pack',   zone: 'B1', t: '14:23:31', ok: false, reason: 'Apparel-only zone' },
  ]);
  const [target, setTarget] = React.useState('A2');

  const simulate = (ok=true) => {
    const items = [['LR-7711','Wool Beanie · Charcoal','A2'],['LR-2204','Canvas Tote · Olive','A2'],['LR-9034','Trail Runner GTX','A3']];
    const r = items[Math.floor(Math.random()*items.length)];
    const t = new Date(); const ts = `${t.getHours()}:${String(t.getMinutes()).padStart(2,'0')}:${String(t.getSeconds()).padStart(2,'0')}`;
    setScans(s => [{ sku:r[0], name:r[1], zone: ok ? target : 'B4', t: ts, ok, reason: ok?null:'Cold storage requires perishable category' }, ...s].slice(0,8));
  };

  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Scanner</div>
          <div className="page-sub">Handheld · device #ZB-03 · paired with Maya R. · zone-locked to <strong>{target}</strong></div>
        </div>
        <div className="page-actions">
          <button className="btn"><Icon d={I.cog} /> Calibrate</button>
          <button className="btn primary" onClick={()=>simulate(true)}><Icon d={I.scan} /> Simulate scan</button>
        </div>
      </div>

      <div style={{display:'grid', gridTemplateColumns:'320px 1fr', gap:24, alignItems:'start'}}>
        {/* Phone preview */}
        <div className="phone">
          <div className="phone-notch" />
          <div className="phone-screen">
            <div style={{display:'flex',alignItems:'center',gap:8,padding:'14px 16px 6px',justifyContent:'space-between'}}>
              <div style={{fontSize:12,fontFamily:'var(--mono)',color:'var(--text-2)'}}>14:24</div>
              <div style={{display:'flex',gap:4,alignItems:'center'}}>
                <span style={{width:14,height:8,border:'1px solid var(--text-2)',borderRadius:1,position:'relative'}}><span style={{position:'absolute',inset:1,right:'25%',background:'var(--text)'}}/></span>
              </div>
            </div>
            <div className="scanner-status">
              <div style={{width:36,height:36,borderRadius:10,background:'var(--accent-soft)',display:'grid',placeItems:'center',color:'var(--accent-text)'}}>
                <Icon d={I.scan} size={18} />
              </div>
              <div style={{flex:1,minWidth:0}}>
                <div style={{fontWeight:600,fontSize:13}}>Place item to {target}</div>
                <div style={{fontSize:11,color:'var(--text-2)'}}>Apparel · slot 12 of 28</div>
              </div>
              <span className="chip success" style={{fontSize:10}}><span className="dot"/>Online</span>
            </div>
            <div className="scan-cam">
              <div className="scan-line" />
              <div className="scan-target">SKU · BARCODE 128</div>
              <div style={{position:'absolute',bottom:8,left:0,right:0,textAlign:'center',color:'rgba(255,255,255,0.7)',fontSize:10,fontFamily:'var(--mono)'}}>aim at barcode · auto-focus</div>
            </div>
            <div style={{padding:'4px 14px 8px',display:'flex',justifyContent:'space-between',alignItems:'center'}}>
              <div style={{fontSize:11,color:'var(--text-3)',fontFamily:'var(--mono)'}}>RECENT</div>
              <div style={{fontSize:11,color:'var(--text-3)'}}>{scans.length} scans</div>
            </div>
            <div style={{padding:'0 12px 12px',display:'flex',flexDirection:'column',gap:6,overflow:'auto',flex:1}}>
              {scans.slice(0,4).map((s,i) => (
                <div key={i} style={{
                  border:'1px solid var(--border)', borderRadius:10, padding:'8px 10px',
                  display:'flex',gap:8,alignItems:'center',
                  background: s.ok ? 'var(--bg)' : 'var(--danger-soft)'
                }}>
                  <div style={{width:24,height:24,borderRadius:6,display:'grid',placeItems:'center',
                    background: s.ok ? 'var(--success-soft)' : 'var(--danger-soft)',
                    color: s.ok ? 'var(--success)' : 'var(--danger)'}}>
                    <Icon d={s.ok?I.check:I.warn} size={14} />
                  </div>
                  <div style={{flex:1,minWidth:0}}>
                    <div style={{fontSize:11.5,fontWeight:600,whiteSpace:'nowrap',overflow:'hidden',textOverflow:'ellipsis'}}>{s.name}</div>
                    <div style={{fontSize:10.5,color:'var(--text-3)',fontFamily:'var(--mono)'}}>{s.sku} → {s.zone}</div>
                  </div>
                  <div style={{fontSize:10,color:'var(--text-3)',fontFamily:'var(--mono)'}}>{s.t.split(':').slice(0,2).join(':')}</div>
                </div>
              ))}
            </div>
            <div style={{padding:'10px 14px 14px',borderTop:'1px solid var(--border)',display:'flex',gap:8}}>
              <button className="btn" style={{flex:1,justifyContent:'center'}}>Skip</button>
              <button className="btn primary" style={{flex:1,justifyContent:'center'}}>Confirm</button>
            </div>
          </div>
        </div>

        {/* Desktop scan log */}
        <div style={{display:'flex',flexDirection:'column',gap:16}}>
          <div className="grid-3">
            <div className="stat"><div className="stat-label">Scans · this shift</div><div className="stat-value">412</div><div className="stat-meta"><span className="delta-up">+18%</span><span className="muted" style={{color:'var(--text-3)'}}>vs avg shift</span></div></div>
            <div className="stat"><div className="stat-label">Mis-scans blocked</div><div className="stat-value">7</div><div className="stat-meta"><span className="muted" style={{color:'var(--text-3)'}}>3 wrong zone · 4 unknown SKU</span></div></div>
            <div className="stat"><div className="stat-label">Avg. scan time</div><div className="stat-value">1.4s</div><div className="stat-meta"><span className="delta-up">−0.3s</span><span className="muted" style={{color:'var(--text-3)'}}>since calibration</span></div></div>
          </div>

          <div className="card">
            <div className="card-head">
              <div className="card-title">Scan log · live</div>
              <div className="tabs" style={{marginLeft:'auto'}}>
                <div className="tab active">All devices</div>
                <div className="tab">ZB-01</div>
                <div className="tab">ZB-03</div>
                <div className="tab">ZB-07</div>
              </div>
            </div>
            <table className="table">
              <thead>
                <tr><th>Time</th><th>SKU</th><th>Item</th><th>From</th><th>To</th><th>Device</th><th>Result</th></tr>
              </thead>
              <tbody>
                {scans.map((s,i) => (
                  <tr key={i}>
                    <td className="num">{s.t}</td>
                    <td className="num">{s.sku}</td>
                    <td>{s.name}</td>
                    <td className="num">{i%2?'C1':'A1'}</td>
                    <td className="num">{s.zone}</td>
                    <td className="num">ZB-0{(i%3)+1}</td>
                    <td>
                      {s.ok
                        ? <span className="chip success"><span className="dot"/>Placed</span>
                        : <span className="chip danger" title={s.reason}><span className="dot"/>Blocked</span>}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="card">
            <div className="card-head"><div className="card-title">Validation rules</div><span className="chip">12 active</span></div>
            <div className="card-body" style={{display:'flex',flexDirection:'column',gap:8}}>
              <RuleRow label="Reject items into zones outside their category" on />
              <RuleRow label="Require photo when item enters Returns (B2)" on />
              <RuleRow label="Block placement when zone is >95% full" on />
              <RuleRow label="Notify owner when SKU dwell exceeds 14 days" on />
              <RuleRow label="Require dual scan for items > $500" />
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
function RuleRow({ label, on }) {
  const [v, setV] = React.useState(!!on);
  return (
    <div style={{display:'flex',alignItems:'center',gap:12,padding:'8px 0',borderBottom:'1px dashed var(--border)'}}>
      <div className={`toggle ${v?'on':''}`} onClick={()=>setV(!v)} />
      <div style={{flex:1, fontSize:13}}>{label}</div>
      <span className="chip">edit</span>
    </div>
  );
}

// ============ INVENTORY ============
function Inventory() {
  const [filter, setFilter] = React.useState('all');
  const items = Dx.items.filter(i => filter==='all' || i.status===filter || (filter==='high' && i.velocity==='high'));
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Inventory</div>
          <div className="page-sub">2,557 items · 184 SKUs · synced with Shopify, Square POS</div>
        </div>
        <div className="page-actions">
          <button className="btn"><Icon d={I.filter} /> Filters</button>
          <button className="btn">Export CSV</button>
          <button className="btn primary"><Icon d={I.plus} /> Add SKU</button>
        </div>
      </div>

      <div className="tabs" style={{alignSelf:'flex-start'}}>
        {[['all','All'],['in-stock','In stock'],['low','Low'],['out','Out'],['high','High velocity']].map(([k,l]) => (
          <div key={k} className={`tab ${filter===k?'active':''}`} onClick={()=>setFilter(k)}>{l}</div>
        ))}
      </div>

      <div className="card" style={{padding:0,overflow:'hidden'}}>
        <table className="table">
          <thead>
            <tr><th>SKU</th><th>Item</th><th>Zone</th><th>Qty</th><th>Price</th><th>Velocity</th><th>Status</th><th>Last scan</th><th></th></tr>
          </thead>
          <tbody>
            {items.map(i => (
              <tr key={i.sku}>
                <td className="num">{i.sku}</td>
                <td><div style={{display:'flex',gap:10,alignItems:'center'}}>
                  <div style={{width:32,height:32,borderRadius:6,background:'var(--panel)',border:'1px solid var(--border)'}} />
                  <span style={{fontWeight:500}}>{i.name}</span>
                </div></td>
                <td><span className="chip accent"><span className="dot"/>{i.zone}</span></td>
                <td className="num">{i.qty}</td>
                <td className="num">${i.price}</td>
                <td>
                  {i.velocity==='high' && <span className="chip success"><span className="dot"/>High</span>}
                  {i.velocity==='medium' && <span className="chip"><span className="dot"/>Medium</span>}
                  {i.velocity==='low' && <span className="chip"><span className="dot"/>Low</span>}
                </td>
                <td>
                  {i.status==='in-stock' && <span className="chip success"><span className="dot"/>In stock</span>}
                  {i.status==='low' && <span className="chip warn"><span className="dot"/>Low</span>}
                  {i.status==='out' && <span className="chip danger"><span className="dot"/>Out</span>}
                </td>
                <td className="num">{i.updated}</td>
                <td><button className="icon-btn" style={{width:28,height:28}}><Icon d={I.more} size={14}/></button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  );
}

// ============ TASKS ============
function Tasks() {
  const cols = ['todo','in-progress','review','done'];
  const colNames = { todo:'To do', 'in-progress':'In progress', review:'Review', done:'Done' };
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Tasks</div>
          <div className="page-sub">Real-time · 6 open · synced with delivery schedule (next: PO #4421 tomorrow 9am)</div>
        </div>
        <div className="page-actions">
          <button className="btn"><Icon d={I.filter} /> Mine</button>
          <button className="btn primary"><Icon d={I.plus} /> New task</button>
        </div>
      </div>

      <div style={{display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:12}}>
        {cols.map(c => (
          <div key={c} className="col">
            <div className="col-head">
              <span>{colNames[c]}</span>
              <span className="count">{Dx.tasks.filter(t=>t.status===c).length}</span>
            </div>
            {Dx.tasks.filter(t=>t.status===c).map(t => (
              <div key={t.id} className="kcard">
                <div className="kmeta"><span className="ktag">{t.id}</span><span className="ktag">·</span><span>{t.zone}</span></div>
                <div className="ktitle">{t.title}</div>
                <div style={{display:'flex',gap:6,alignItems:'center',justifyContent:'space-between'}}>
                  <PriorityChip p={t.priority} />
                  <span style={{fontSize:11,color:'var(--text-3)'}}><Icon d={I.clock} size={11}/> {t.due}</span>
                </div>
                <div style={{display:'flex',alignItems:'center',gap:6,marginTop:4}}>
                  <div className="avatar" style={{width:20,height:20,fontSize:10}}>{t.assignee.split(' ').map(p=>p[0]).join('')}</div>
                  <span style={{fontSize:11,color:'var(--text-2)'}}>{t.assignee}</span>
                </div>
              </div>
            ))}
            <button className="btn ghost" style={{marginTop:'auto',justifyContent:'center'}}><Icon d={I.plus} size={12}/> Add</button>
          </div>
        ))}
      </div>
    </>
  );
}

// ============ PIPELINES ============
function Pipelines() {
  const p = Dx.pipelines['item-restoration'];
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Pipelines</div>
          <div className="page-sub">Custom workflows for non-standard item flows · 3 active pipelines</div>
        </div>
        <div className="page-actions">
          <div className="role-pill">
            <span className="swatch" />
            <select defaultValue="restoration">
              <option value="restoration">Item Restoration · 8 active</option>
              <option>Returns Triage · 14 active</option>
              <option>B2B Outbound · 6 active</option>
            </select>
          </div>
          <button className="btn"><Icon d={I.cog}/> Edit stages</button>
          <button className="btn primary"><Icon d={I.plus}/> New pipeline</button>
        </div>
      </div>

      <div className="kanban">
        {p.stages.map((stage, si) => (
          <div className="col" key={si}>
            <div className="col-head">
              <span>{si+1}. {stage}</span>
              <span className="count">{p.cards.filter(c=>c.stage===si).length}</span>
            </div>
            {p.cards.filter(c=>c.stage===si).map(c => (
              <div className="kcard" key={c.id}>
                <div className="kmeta"><span className="ktag">{c.id}</span><span className="ktag">{c.sku}</span></div>
                <div className="ktitle">{c.title}</div>
                <div style={{display:'flex',justifyContent:'space-between',alignItems:'center',marginTop:2}}>
                  <PriorityChip p={c.priority} />
                  <div style={{display:'flex',alignItems:'center',gap:8}}>
                    <span style={{fontSize:10,color:'var(--text-3)',fontFamily:'var(--mono)'}}>{c.age}</span>
                    <div className="avatar" style={{width:20,height:20,fontSize:10}}>{c.owner}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ))}
      </div>

      <div className="grid-3">
        <div className="card"><div className="card-head"><div className="card-title">Stage SLAs</div></div><div className="card-body" style={{display:'flex',flexDirection:'column',gap:8}}>
          {[['Intake','24h','99%'],['Triage','2d','94%'],['Repair','5d','81%'],['QA','1d','97%']].map(r => (
            <div key={r[0]} style={{display:'flex',justifyContent:'space-between',padding:'6px 0',borderBottom:'1px dashed var(--border)',fontSize:13}}>
              <span>{r[0]}</span><span className="muted">{r[1]}</span><span style={{fontFamily:'var(--mono)'}}>{r[2]}</span>
            </div>
          ))}
        </div></div>
        <div className="card"><div className="card-head"><div className="card-title">Bottleneck</div><span className="chip warn"><span className="dot"/>Repair stage</span></div><div className="card-body" style={{fontSize:13,lineHeight:1.55}}>
          <div>Cast Iron · re-season aging 3d (above 5d SLA threshold). Two items waiting on Maya R.</div>
          <div style={{marginTop:10,display:'flex',gap:6}}>
            <button className="btn">View load</button><button className="btn primary">Reassign</button>
          </div>
        </div></div>
        <div className="card"><div className="card-head"><div className="card-title">Throughput · 30d</div></div><div className="card-body">
          <Sparkline data={[6,8,7,9,11,10,12,14,13,16,15,17,19,18,20,22,21,23,25,24,27,29,28,30,32,31,33,35,34,36]} positive />
          <div style={{display:'flex',justifyContent:'space-between',marginTop:8,fontSize:12}}>
            <span className="muted" style={{color:'var(--text-3)'}}>Apr 9</span>
            <strong>+38% restored</strong>
            <span className="muted" style={{color:'var(--text-3)'}}>May 9</span>
          </div>
        </div></div>
      </div>
    </>
  );
}

Object.assign(window, { Scanner, Inventory, Tasks, Pipelines });
