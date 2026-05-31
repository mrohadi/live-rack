/* global React */

const USERS = [
  { id:1, name:'Avery Chen',    title:'Ops Manager',    email:'avery@live-rack.co',   role:'Admin',     zones:'All',          shift:'Day',     status:'active',  last:'just now',  initials:'AC', tone:'#2563eb' },
  { id:2, name:'Maya Reyes',    title:'Floor Lead',     email:'maya@live-rack.co',    role:'Manager',   zones:'A2, A3, B1',   shift:'Day',     status:'active',  last:'2m ago',    initials:'MR', tone:'#7c3aed' },
  { id:3, name:'Theo Kim',      title:'Analyst',        email:'theo@live-rack.co',    role:'Manager',   zones:'All',          shift:'Day',     status:'active',  last:'14m ago',   initials:'TK', tone:'#0891b2' },
  { id:4, name:'Jordan Park',   title:'Receiving',      email:'jordan@live-rack.co',  role:'Staff',     zones:'A1, C1',       shift:'Open',    status:'active',  last:'8m ago',    initials:'JP', tone:'#16a34a' },
  { id:5, name:'Aniyah Brooks', title:'Showroom Lead',  email:'aniyah@live-rack.co',  role:'Staff',     zones:'C2',           shift:'Day',     status:'active',  last:'1h ago',    initials:'AB', tone:'#d97706' },
  { id:6, name:'Sam Okafor',    title:'Night Restock',  email:'sam@live-rack.co',     role:'Staff',     zones:'B1, C1',       shift:'Night',   status:'off',     last:'8h ago',    initials:'SO', tone:'#5b6577' },
  { id:7, name:'Priya Nair',    title:'Returns',        email:'priya@live-rack.co',   role:'Staff',     zones:'B2',           shift:'Day',     status:'active',  last:'24m ago',   initials:'PN', tone:'#dc2626' },
  { id:8, name:'Luca Romano',   title:'Repair Tech',    email:'luca@live-rack.co',    role:'Staff',     zones:'B2 pipeline',  shift:'Day',     status:'break',   last:'3m ago',    initials:'LR', tone:'#0b1220' },
  { id:9, name:'Devon Hayes',   title:'Forklift',       email:'devon@live-rack.co',   role:'Staff',     zones:'C1, C3',       shift:'Open',    status:'active',  last:'just now',  initials:'DH', tone:'#16a34a' },
  { id:10,name:'Riya Shah',     title:'POS Lead',       email:'riya@live-rack.co',    role:'Staff',     zones:'C2',           shift:'Day',     status:'active',  last:'5m ago',    initials:'RS', tone:'#7c3aed' },
  { id:11,name:'Owen Walsh',    title:'Auditor',        email:'owen@audit.io',        role:'Read-only', zones:'All',          shift:'On-call', status:'active',  last:'2d ago',    initials:'OW', tone:'#5b6577' },
  { id:12,name:'API · Shopify', title:'Service token',  email:'svc-shopify@…',       role:'Service',   zones:'—',            shift:'24/7',    status:'active',  last:'live',      initials:'SH', tone:'#16a34a' },
];

const ROLE_ROWS = [
  { role:'Admin',     people:1, perms:'Full · settings, billing, integrations, users' },
  { role:'Manager',   people:2, perms:'Zones, tasks, pipelines, analytics, scanner rules' },
  { role:'Staff',     people:7, perms:'Scanner, assigned zones, own tasks' },
  { role:'Read-only', people:1, perms:'View dashboards & exports' },
  { role:'Service',   people:1, perms:'API only · scoped tokens' },
];

const PERMS = [
  ['View dashboards',      true,  true,  true,  true ],
  ['Edit zones & layout',  true,  true,  false, false],
  ['Approve mis-scans',    true,  true,  false, false],
  ['Manage pipelines',     true,  true,  false, false],
  ['Run scanner',          true,  true,  true,  false],
  ['Move inventory',       true,  true,  true,  false],
  ['Manage tasks (any)',   true,  true,  false, false],
  ['Manage tasks (own)',   true,  true,  true,  false],
  ['Edit users',           true,  false, false, false],
  ['Manage integrations',  true,  false, false, false],
  ['Export reports',       true,  true,  false, true ],
];

function Users() {
  const [filter, setFilter] = React.useState('all');
  const [selected, setSelected] = React.useState(2);
  const filtered = USERS.filter(u => filter==='all' || u.role.toLowerCase()===filter);
  const sel = USERS.find(u => u.id === selected);

  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Users & Access</div>
          <div className="page-sub">12 members · 5 roles · last permission change 3d ago by Avery Chen</div>
        </div>
        <div className="page-actions">
          <button className="btn"><Icon d={I.cog}/> Audit log</button>
          <button className="btn">Invite link</button>
          <button className="btn primary"><Icon d={I.plus}/> Add user</button>
        </div>
      </div>

      <div className="grid-3">
        <div className="stat"><div className="stat-label">Active now</div><div className="stat-value">7</div><div className="stat-meta"><span className="muted" style={{color:'var(--text-3)'}}>on floor · scanning</span></div></div>
        <div className="stat"><div className="stat-label">Pending invites</div><div className="stat-value">2</div><div className="stat-meta"><span className="muted" style={{color:'var(--text-3)'}}>expires in 6 days</span></div></div>
        <div className="stat"><div className="stat-label">2FA coverage</div><div className="stat-value">92%</div><div className="stat-meta"><span className="delta-up">+8pp</span><span className="muted" style={{color:'var(--text-3)'}}>since policy</span></div></div>
      </div>

      <div style={{display:'grid', gridTemplateColumns:'1fr 320px', gap:16, alignItems:'start'}}>
        <div style={{display:'flex',flexDirection:'column',gap:14}}>
          <div className="tabs" style={{alignSelf:'flex-start'}}>
            {[['all','All · 12'],['admin','Admin'],['manager','Manager'],['staff','Staff'],['read-only','Read-only'],['service','Service']].map(([k,l])=>(
              <div key={k} className={`tab ${filter===k?'active':''}`} onClick={()=>setFilter(k)}>{l}</div>
            ))}
          </div>

          <div className="card" style={{padding:0,overflow:'hidden'}}>
            <table className="table">
              <thead>
                <tr><th>Member</th><th>Role</th><th>Zones</th><th>Shift</th><th>Status</th><th>Last seen</th><th></th></tr>
              </thead>
              <tbody>
                {filtered.map(u => (
                  <tr key={u.id} onClick={()=>setSelected(u.id)} style={{cursor:'pointer', background: selected===u.id ? 'var(--panel)' : undefined}}>
                    <td><div style={{display:'flex',alignItems:'center',gap:10}}>
                      <div className="avatar" style={{background:`linear-gradient(135deg, ${u.tone}, color-mix(in oklab, ${u.tone} 60%, #000))`}}>{u.initials}</div>
                      <div>
                        <div style={{fontSize:13, fontWeight:600}}>{u.name}</div>
                        <div style={{fontSize:11.5, color:'var(--text-3)'}}>{u.title} · {u.email}</div>
                      </div>
                    </div></td>
                    <td><RoleChip role={u.role}/></td>
                    <td className="num">{u.zones}</td>
                    <td>{u.shift}</td>
                    <td>
                      {u.status==='active' && <span className="chip success"><span className="dot"/>Active</span>}
                      {u.status==='break'  && <span className="chip warn"><span className="dot"/>On break</span>}
                      {u.status==='off'    && <span className="chip"><span className="dot"/>Off</span>}
                    </td>
                    <td className="num">{u.last}</td>
                    <td><button className="icon-btn" style={{width:28,height:28}} onClick={e=>e.stopPropagation()}><Icon d={I.more} size={14}/></button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="card">
            <div className="card-head">
              <div className="card-title">Role permissions</div>
              <span className="chip">5 roles</span>
            </div>
            <div className="card-body" style={{padding:0}}>
              <table className="table">
                <thead>
                  <tr><th>Permission</th><th style={{textAlign:'center'}}>Admin</th><th style={{textAlign:'center'}}>Manager</th><th style={{textAlign:'center'}}>Staff</th><th style={{textAlign:'center'}}>Read-only</th></tr>
                </thead>
                <tbody>
                  {PERMS.map((p,i) => (
                    <tr key={i}>
                      <td>{p[0]}</td>
                      {[1,2,3,4].map(c => (
                        <td key={c} style={{textAlign:'center'}}>
                          {p[c] ? <Icon d={I.check} size={14} stroke={2.2}/> : <span style={{color:'var(--text-3)'}}>—</span>}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Side detail panel */}
        <div className="card" style={{padding:0, position:'sticky', top:0}}>
          <div className="card-head" style={{flexDirection:'column', alignItems:'flex-start', gap:10, padding:'18px 16px'}}>
            <div style={{display:'flex', alignItems:'center', gap:12, width:'100%'}}>
              <div className="avatar" style={{width:44,height:44,fontSize:15,background:`linear-gradient(135deg, ${sel.tone}, color-mix(in oklab, ${sel.tone} 60%, #000))`}}>{sel.initials}</div>
              <div style={{flex:1, minWidth:0}}>
                <div style={{fontSize:15, fontWeight:600}}>{sel.name}</div>
                <div style={{fontSize:12, color:'var(--text-2)'}}>{sel.title}</div>
              </div>
              <span className={`chip ${sel.status==='active'?'success':sel.status==='break'?'warn':''}`}>
                <span className="dot"/>{sel.status}
              </span>
            </div>
          </div>
          <div className="card-body" style={{paddingTop:8}}>
            <div className="kv"><span className="k">Email</span><span className="v" style={{fontFamily:'var(--mono)', fontSize:12}}>{sel.email}</span></div>
            <div className="kv"><span className="k">Role</span><span className="v"><RoleChip role={sel.role}/></span></div>
            <div className="kv"><span className="k">Zone access</span><span className="v">{sel.zones}</span></div>
            <div className="kv"><span className="k">Shift</span><span className="v">{sel.shift}</span></div>
            <div className="kv"><span className="k">Last seen</span><span className="v" style={{fontFamily:'var(--mono)',fontSize:12}}>{sel.last}</span></div>
            <div className="kv"><span className="k">2FA</span><span className="v"><span className="chip success"><span className="dot"/>Enabled</span></span></div>

            <div style={{marginTop:14, fontSize:11, fontWeight:600, color:'var(--text-3)', letterSpacing:'0.06em'}}>RECENT ACTIVITY</div>
            <div style={{display:'flex', flexDirection:'column', gap:0, marginTop:6}}>
              {[
                ['Scanned 12 items into A2', '14:18', 'accent'],
                ['Closed task T-208',         '13:54', 'success'],
                ['Approved restock plan',     '13:40', 'violet'],
                ['Sign-in · scanner ZB-03',   '08:02', 'muted'],
              ].map((r,i)=>(
                <div key={i} className="activity-row">
                  <div className="activity-dot" style={{background: r[2]==='success'?'var(--success)':r[2]==='violet'?'var(--violet)':r[2]==='muted'?'var(--text-3)':'var(--accent)'}}/>
                  <div className="activity-text">{r[0]}</div>
                  <div className="activity-meta">{r[1]}</div>
                </div>
              ))}
            </div>
          </div>
          <div className="card-foot">
            <button className="btn ghost">Reset password</button>
            <button className="btn primary">Edit access</button>
          </div>
        </div>
      </div>
    </>
  );
}

function RoleChip({ role }) {
  const map = {
    'Admin': 'danger', 'Manager': 'accent', 'Staff': 'success',
    'Read-only': '', 'Service': 'violet'
  };
  return <span className={`chip ${map[role]||''}`}><span className="dot"/>{role}</span>;
}

Object.assign(window, { Users });
