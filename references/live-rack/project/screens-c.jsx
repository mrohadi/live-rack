/* global React, LR_DATA */
const Dy = window.LR_DATA;

// ============ ANALYTICS ============
function Analytics({ accent }) {
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Analytics</div>
          <div className="page-sub">Cross-signal · sales × zone × demographic × weather × transit</div>
        </div>
        <div className="page-actions">
          <div className="tabs">
            <div className="tab">24h</div>
            <div className="tab active">7d</div>
            <div className="tab">30d</div>
            <div className="tab">QTD</div>
          </div>
          <button className="btn"><Icon d={I.filter} /> Segment</button>
          <button className="btn">Export</button>
        </div>
      </div>

      <div className="stat-grid">
        <Stat label="GMV · 7d"        value="$184.2k" delta="+9.4%"  up meta="vs prev 7d"   spark={[120,128,134,142,151,162,178,184]} />
        <Stat label="Time-to-sell"    value="3.4d"   delta="−0.6d"  up meta="median"         spark={[5.0,4.7,4.4,4.1,3.9,3.6,3.5,3.4]} />
        <Stat label="Sell-through"    value="62%"    delta="+4pp"   up meta="apparel-led"    spark={[51,53,55,56,58,59,61,62]} />
        <Stat label="Attach rate"     value="1.42"   delta="+0.08"  up meta="items per ticket" spark={[1.30,1.31,1.34,1.36,1.38,1.40,1.41,1.42]} />
      </div>

      <div className="grid-2-1">
        <div className="card">
          <div className="card-head">
            <div className="card-title">Zone performance</div>
            <span className="chip violet"><span className="dot"/>Showroom outperforming Bulk 6.2×</span>
          </div>
          <div className="card-body">
            {Dy.zonePerf.map(z => (
              <div className="bar-row" key={z.zone}>
                <span className="bar-label">{z.zone}</span>
                <div className="bar-track"><div className="bar-fill" style={{width: `${(z.sales/18420)*100}%`, background: accent}}/></div>
                <span className="bar-value">${(z.sales/1000).toFixed(1)}k</span>
                <span className={z.change>=0?'delta-up':'delta-down'} style={{width:48,textAlign:'right',fontSize:11.5,fontFamily:'var(--mono)'}}>{z.change>=0?'+':''}{z.change}%</span>
              </div>
            ))}
            <div style={{marginTop:14,fontSize:11.5,color:'var(--text-3)',fontFamily:'var(--mono)'}}>$ / sq ft · last 7 days</div>
          </div>
        </div>

        <div className="card">
          <div className="card-head"><div className="card-title">Customer mix</div></div>
          <div className="card-body" style={{display:'flex',flexDirection:'column',gap:14}}>
            <div>
              <div style={{fontSize:11.5,color:'var(--text-3)',marginBottom:6}}>AGE</div>
              {Dy.demographics.map(d => (
                <div className="bar-row" key={d.label} style={{padding:'3px 0'}}>
                  <span className="bar-label" style={{width:40,fontSize:11.5}}>{d.label}</span>
                  <div className="bar-track" style={{height:6}}><div className="bar-fill" style={{width:`${d.pct*3}%`,background:accent,height:6}}/></div>
                  <span className="bar-value" style={{width:36,fontSize:11}}>{d.pct}%</span>
                </div>
              ))}
            </div>
            <div>
              <div style={{fontSize:11.5,color:'var(--text-3)',marginBottom:6}}>WITHIN 5KM</div>
              <div className="bar-row" style={{padding:'3px 0'}}><span className="bar-label" style={{width:80,fontSize:11.5}}>Residents</span><div className="bar-track" style={{height:6}}><div className="bar-fill" style={{width:'58%',background:accent,height:6}}/></div><span className="bar-value" style={{width:36}}>58%</span></div>
              <div className="bar-row" style={{padding:'3px 0'}}><span className="bar-label" style={{width:80,fontSize:11.5}}>Commuters</span><div className="bar-track" style={{height:6}}><div className="bar-fill" style={{width:'27%',background:accent,height:6}}/></div><span className="bar-value" style={{width:36}}>27%</span></div>
              <div className="bar-row" style={{padding:'3px 0'}}><span className="bar-label" style={{width:80,fontSize:11.5}}>Visitors</span><div className="bar-track" style={{height:6}}><div className="bar-fill" style={{width:'15%',background:accent,height:6}}/></div><span className="bar-value" style={{width:36}}>15%</span></div>
            </div>
          </div>
        </div>
      </div>

      <div className="grid-2">
        <div className="card">
          <div className="card-head"><div className="card-title">Frequently sold together</div><span className="chip">last 30d</span></div>
          <div className="card-body" style={{padding:0}}>
            <table className="table">
              <thead><tr><th>Item A</th><th>Item B</th><th>Co-purchase lift</th><th>Suggested action</th></tr></thead>
              <tbody>
                {Dy.combos.map((c,i) => (
                  <tr key={i}>
                    <td>{c.a}</td><td>{c.b}</td>
                    <td className="num"><span className="chip success"><span className="dot"/>{c.lift}</span></td>
                    <td className="muted" style={{color:'var(--text-2)'}}>Adjacent placement → {i===0?'A2-front':i===1?'A3-end':'C2-aisle'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="card">
          <div className="card-head"><div className="card-title">External signals</div><span className="chip accent"><span className="dot"/>3 actionable</span></div>
          <div className="card-body" style={{display:'flex',flexDirection:'column',gap:10}}>
            <Signal kind="weather" title="Light rain forecast 16:00–20:00" body="Boost umbrellas, rain shells to A2-front. +12% lift expected." />
            <Signal kind="transit" title="Caltrain 17:42 northbound +8min delay" body="Commuter footfall pushes peak to 18:10. Hold restock until 18:30." />
            <Signal kind="event"   title="Civic Center concert · 19:30 doors" body="Expect +400 walk-bys. Move show-stoppers to C2-window." />
            <Signal kind="weather" title="Heat advisory tomorrow · 92°F" body="Reposition cold beverages to staging C3 by 09:00." />
          </div>
        </div>
      </div>
    </>
  );
}
function Signal({ kind, title, body }) {
  const icon = kind==='weather' ? '☂' : kind==='transit' ? '⇆' : '★';
  const tone = kind==='weather' ? 'accent' : kind==='transit' ? 'violet' : 'warn';
  return (
    <div style={{display:'flex',gap:12,padding:'10px',border:'1px solid var(--border)',borderRadius:8}}>
      <div style={{width:32,height:32,borderRadius:8,display:'grid',placeItems:'center',
        background: tone==='accent'?'var(--accent-soft)':tone==='violet'?'var(--violet-soft)':'var(--warning-soft)',
        color: tone==='accent'?'var(--accent-text)':tone==='violet'?'var(--violet)':'var(--warning)',
        fontSize:16,flex:'none'}}>{icon}</div>
      <div style={{flex:1,minWidth:0}}>
        <div style={{fontSize:13,fontWeight:600}}>{title}</div>
        <div style={{fontSize:12,color:'var(--text-2)',marginTop:3,lineHeight:1.45}}>{body}</div>
      </div>
      <button className="btn" style={{alignSelf:'center'}}>Apply</button>
    </div>
  );
}

// ============ INTEGRATIONS ============
function Integrations() {
  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Integrations</div>
          <div className="page-sub">7 connected · 2 available · last sync 12 seconds ago</div>
        </div>
        <div className="page-actions">
          <button className="btn"><Icon d={I.cog}/> Webhooks</button>
          <button className="btn primary"><Icon d={I.plus}/> Add integration</button>
        </div>
      </div>

      <div className="grid-3">
        <div className="card" style={{padding:16,gap:8,display:'flex',flexDirection:'column'}}>
          <div className="card-title">Sync status</div>
          {[['Shopify',98,'12s'],['Square POS',100,'live'],['Stripe',99,'1m'],['Shippo',96,'4m']].map(r => (
            <div key={r[0]} style={{display:'flex',alignItems:'center',gap:10,fontSize:12.5}}>
              <span style={{flex:1}}>{r[0]}</span>
              <div style={{width:120,height:6,background:'var(--panel)',borderRadius:4,overflow:'hidden'}}>
                <div style={{width:`${r[1]}%`,height:'100%',background:'var(--success)'}}/>
              </div>
              <span style={{width:36,textAlign:'right',fontFamily:'var(--mono)',fontSize:11,color:'var(--text-3)'}}>{r[2]}</span>
            </div>
          ))}
        </div>
        <div className="card" style={{padding:16,display:'flex',flexDirection:'column',gap:8}}>
          <div className="card-title">Events · last hour</div>
          <div style={{fontSize:12,color:'var(--text-2)',display:'flex',flexDirection:'column',gap:4,fontFamily:'var(--mono)'}}>
            <div>14:24:08 · square.sale · LR-9034 → settled</div>
            <div>14:21:11 · shopify.order.create · #10241</div>
            <div>14:18:44 · shippo.label.created · 1Z…48F</div>
            <div>14:09:02 · shopify.inventory.delta · −3 SKUs</div>
            <div>13:58:21 · openweather.forecast · rain @16:00</div>
          </div>
        </div>
        <div className="card" style={{padding:16,display:'flex',flexDirection:'column',gap:10}}>
          <div className="card-title">Outbound visibility</div>
          <div style={{fontSize:13,color:'var(--text-2)'}}>Item locations expose to:</div>
          <div style={{display:'flex',flexWrap:'wrap',gap:6}}>
            <span className="chip success"><span className="dot"/>Shopify storefront</span>
            <span className="chip success"><span className="dot"/>Curbside app</span>
            <span className="chip success"><span className="dot"/>3PL Shippo</span>
            <span className="chip"><span className="dot"/>Wholesale (off)</span>
          </div>
          <div style={{marginTop:'auto',display:'flex',gap:6}}>
            <button className="btn" style={{flex:1,justifyContent:'center'}}>Configure</button>
          </div>
        </div>
      </div>

      <div className="int-grid">
        {Dy.integrations.map(i => (
          <div className="int-card" key={i.name}>
            <div className="int-head">
              <div className="int-logo" style={{background: i.bg}}>{i.initials}</div>
              <div style={{flex:1}}>
                <div className="int-name">{i.name}</div>
                <div className="int-cat">{i.cat}</div>
              </div>
              {i.status==='connected' && <span className="chip success"><span className="dot"/>Connected</span>}
              {i.status==='paused' && <span className="chip warn"><span className="dot"/>Paused</span>}
              {i.status==='available' && <span className="chip"><span className="dot"/>Available</span>}
            </div>
            <div style={{display:'flex',justifyContent:'space-between',fontSize:12,color:'var(--text-2)',borderTop:'1px dashed var(--border)',paddingTop:10}}>
              <span>Last sync</span><span style={{fontFamily:'var(--mono)'}}>{i.last}</span>
            </div>
            <div style={{display:'flex',gap:6}}>
              <button className="btn" style={{flex:1,justifyContent:'center'}}>{i.status==='available'?'Connect':'Manage'}</button>
              <button className="icon-btn"><Icon d={I.more}/></button>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

Object.assign(window, { Analytics, Integrations });
