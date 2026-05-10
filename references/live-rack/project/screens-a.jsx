/* global React, LR_DATA */
const D = window.LR_DATA;

// ============ DASHBOARD ============
function Dashboard({ accent }) {
  const [tick, setTick] = React.useState(0);
  React.useEffect(() => {const t = setInterval(() => setTick((x) => x + 1), 4000);return () => clearInterval(t);}, []);

  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Good afternoon, Avery</div>
          <div className="page-sub">Store #14 · Wed May 9, 2026 · 14:24 PT · 7 staff on floor</div>
        </div>
        <div className="page-actions">
          <button className="btn"><Icon d={I.filter} /> Today</button>
          <button className="btn primary"><Icon d={I.plus} /> New task</button>
        </div>
      </div>

      <div className="stat-grid">
        <Stat label="Sales today" value="$28,420" delta="+12.4%" up meta="vs. yesterday" spark={D.sales24} />
        <Stat label="Items moved" value="1,284" delta="+8.1%" up meta="scans · 24h" spark={D.sales24.map((v) => v * 0.6 + 8)} />
        <Stat label="Misplaced (open)" value="14" delta="-3" up meta="auto-flagged" spark={[8, 9, 11, 12, 15, 18, 17, 16, 14, 15, 16, 15, 14]} negativeIsGood />
        <Stat label="Fulfillment SLA" value="98.6%" delta="+0.4pp" up meta="rolling 7d" spark={[97, 97.5, 97, 98, 98.4, 98.6, 98.6]} />
      </div>

      <div className="grid-2-1">
        <div className="card">
          <div className="card-head">
            <div className="card-title">Live activity</div>
            <span className="chip success"><span className="dot" /> Connected · 7 sources</span>
            <div style={{ marginLeft: 'auto' }} className="muted" style={{ marginLeft: 'auto', fontSize: 12, color: 'var(--text-3)' }}>Updated {tick}s ago</div>
          </div>
          <div className="card-body" style={{ paddingTop: 6, paddingBottom: 6 }}>
            <div className="activity">
              {D.activity.map((a, i) =>
              <div className="activity-row" key={i}>
                  <div className="activity-dot" style={{ background:
                  a.kind === 'success' ? 'var(--success)' : a.kind === 'warn' ? 'var(--warning)' : a.kind === 'violet' ? 'var(--violet)' : a.kind === 'muted' ? 'var(--text-3)' : 'var(--accent)' }} />
                  <div className="activity-text">
                    <span className="muted" style={{ color: 'var(--text-2)' }}>{a.who}</span> {a.text}<strong>{a.em}</strong>{a.tail}
                  </div>
                  <div className="activity-meta">{a.t}</div>
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="card">
          <div className="card-head">
            <div className="card-title">Today's tasks</div>
            <span className="chip">6 open</span>
          </div>
          <div className="card-body" style={{ display: 'flex', flexDirection: 'column', gap: 10, paddingTop: 10 }}>
            {D.tasks.slice(0, 5).map((t) =>
            <div key={t.id} style={{ display: 'flex', gap: 10, alignItems: 'flex-start', padding: '8px 0', borderBottom: '1px dashed var(--border)' }}>
                <input type="checkbox" defaultChecked={t.status === 'done'} style={{ marginTop: 3 }} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 500, lineHeight: 1.3 }}>{t.title}</div>
                  <div style={{ fontSize: 11.5, color: 'var(--text-3)', marginTop: 3, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                    <span>{t.assignee}</span><span>·</span><span>{t.due}</span><span>·</span><span style={{ fontFamily: 'var(--mono)' }}>{t.zone}</span>
                  </div>
                </div>
                <PriorityChip p={t.priority} />
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="grid-2">
        <div className="card">
          <div className="card-head">
            <div className="card-title">Sales by zone · today</div>
            <div className="tabs" style={{ marginLeft: 'auto' }}>
              <div className="tab active">Today</div><div className="tab">7d</div><div className="tab">30d</div>
            </div>
          </div>
          <div className="card-body">
            {D.zonePerf.map((z) =>
            <div className="bar-row" key={z.zone}>
                <span className="bar-label">{z.zone}</span>
                <div className="bar-track"><div className="bar-fill" style={{ width: `${z.sales / 18420 * 100}%`, background: accent }} /></div>
                <span className="bar-value">${(z.sales / 1000).toFixed(1)}k</span>
                <span className={z.change >= 0 ? 'delta-up' : 'delta-down'} style={{ width: 48, textAlign: 'right', fontSize: 11.5, fontFamily: 'var(--mono)' }}>{z.change >= 0 ? '+' : ''}{z.change}%</span>
              </div>
            )}
          </div>
        </div>

        <div className="card">
          <div className="card-head">
            <div className="card-title">Activity heatmap · last 7 days</div>
            <span className="chip accent"><span className="dot" /> Sat 14:00 peak</span>
          </div>
          <div className="card-body">
            <Heatmap data={D.heatmap} accent={accent} />
          </div>
        </div>
      </div>
    </>);

}

function Stat({ label, value, delta, up, meta, spark, negativeIsGood }) {
  const good = negativeIsGood ? !up : up;
  return (
    <div className="stat">
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      <Sparkline data={spark || []} positive={good} />
      <div className="stat-meta">
        <span className={good ? 'delta-up' : 'delta-down'}>
          <Icon d={good ? I.arrowUp : I.arrowDn} size={12} /> {delta}
        </span>
        <span className="muted" style={{ color: 'var(--text-3)' }}>{meta}</span>
      </div>
    </div>);

}
function Sparkline({ data, positive }) {
  if (!data?.length) return null;
  const max = Math.max(...data),min = Math.min(...data);
  const range = max - min || 1;
  const w = 100,h = 28;
  const pts = data.map((v, i) => `${i / (data.length - 1) * w},${h - (v - min) / range * h}`).join(' ');
  const last = data[data.length - 1];
  const lx = w,ly = h - (last - min) / range * h;
  const color = positive ? 'var(--success)' : 'var(--danger)';
  return (
    <svg className="spark" viewBox={`0 0 ${w} ${h + 2}`} preserveAspectRatio="none" style={{ height: 30 }}>
      <polyline points={pts} fill="none" stroke={color} strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
      <circle cx={lx} cy={ly} r="1.6" fill={color} />
    </svg>);

}
function Heatmap({ data, accent }) {
  const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div className="heat" style={{ gridTemplateColumns: '36px repeat(24,1fr)' }}>
        <div />
        {Array.from({ length: 24 }).map((_, h) =>
        <div key={h} style={{ fontSize: 9.5, color: 'var(--text-3)', textAlign: 'center', fontFamily: 'var(--mono)' }}>{h % 6 === 0 ? h : ''}</div>
        )}
      </div>
      {data.map((row, di) =>
      <div className="heat" key={di}>
          <div className="heat-row-label">{days[di]}</div>
          {row.map((v, h) =>
        <div className="heat-cell" key={h}
        title={`${days[di]} ${h}:00 · ${v * 100 | 0}`}
        style={{ background: `color-mix(in oklab, ${accent} ${Math.round(v * 100)}%, var(--panel))` }} />
        )}
        </div>
      )}
    </div>);

}
function PriorityChip({ p }) {
  const map = { high: ['danger', 'High'], medium: ['warn', 'Med'], low: ['', 'Low'] };
  const [cls, label] = map[p];
  return <span className={`chip ${cls}`}><span className="dot" />{label}</span>;
}

// ============ MAP ============
function MapPage({ accent }) {
  const [selected, setSelected] = React.useState('A2');
  const [view, setView] = React.useState('zones');
  const z = D.zones.find((x) => x.id === selected);

  return (
    <>
      <div className="page-head">
        <div>
          <div className="page-title">Map & Zones</div>
          <div className="page-sub">Floor 1 · 11 zones · 2,557 items tracked · last sync 12s</div>
        </div>
        <div className="page-actions">
          <div className="tabs">
            <div className={`tab ${view === 'zones' ? 'active' : ''}`} onClick={() => setView('zones')}>Zones</div>
            <div className={`tab ${view === 'heat' ? 'active' : ''}`} onClick={() => setView('heat')}>Heat</div>
            <div className={`tab ${view === 'items' ? 'active' : ''}`} onClick={() => setView('items')}>Items</div>
          </div>
          <button className="btn"><Icon d={I.filter} /> Filter</button>
          <button className="btn primary"><Icon d={I.plus} /> New zone</button>
        </div>
      </div>

      <div className="map-shell">
        <div>
          <div className="map-canvas" data-comment-anchor="8b80551c1a-div-195-11">
            {D.zones.map((zn) => {
              const fillPct = zn.items / zn.capacity;
              const heatBg = view === 'heat' ? `color-mix(in oklab, ${accent} ${Math.round(fillPct * 90)}%, var(--panel))` : null;
              return (
                <div key={zn.id} className={`zone ${selected === zn.id ? 'selected' : ''}`}
                onClick={() => setSelected(zn.id)}
                style={{
                  left: `${zn.x}%`, top: `${zn.y}%`, width: `${zn.w}%`, height: `${zn.h}%`,
                  color: zn.color,
                  background: heatBg || undefined
                }}>
                  <div className="zone-name">{zn.name}</div>
                  <div className="zone-meta">{zn.items}/{zn.capacity} · {Math.round(fillPct * 100)}%</div>
                  {view === 'items' && (() => {
                    const cols = Math.max(4, Math.round(zn.w / 4));
                    const rows = Math.max(3, Math.round(zn.h / 6));
                    const total = cols * rows;
                    const filled = Math.min(total, Math.round(total * fillPct));
                    const itemsHere = D.items.filter(it => it.zone === zn.id);
                    return (
                      <div className="zone-rack-grid"
                        style={{ gridTemplateColumns: `repeat(${cols}, 1fr)`, gridTemplateRows: `repeat(${rows}, 1fr)` }}>
                        {Array.from({ length: total }).map((_, i) => {
                          const isFilled = i < filled;
                          const item = isFilled ? itemsHere[i % Math.max(1, itemsHere.length)] : null;
                          const col = String.fromCharCode(65 + (i % cols));
                          const row = Math.floor(i / cols) + 1;
                          const tip = item
                            ? `${zn.id}-${String(row).padStart(2,'0')}${col} · ${item.sku}`
                            : `${zn.id}-${String(row).padStart(2,'0')}${col} · empty`;
                          const cls = item?.status === 'low' ? 'warn' : item?.status === 'out' ? 'out' : '';
                          return (
                            <span key={i}
                              className={`slot ${isFilled?'filled':''} ${cls}`}
                              data-tip={tip}
                              onClick={(e) => e.stopPropagation()} />
                          );
                        })}
                      </div>
                    );
                  })()}
                </div>);

            })}
            {/* SKU location pins for selected zone */}
            {(() => {
              const zn = D.zones.find(x => x.id === selected);
              if (!zn || view !== 'items') return null;
              const cols = Math.max(4, Math.round(zn.w / 4));
              const rows = Math.max(3, Math.round(zn.h / 6));
              const itemsHere = D.items.filter(it => it.zone === zn.id);
              return itemsHere.slice(0, 4).map((it, i) => {
                const idx = i * 3 + 1;
                const cx = (idx % cols) + 0.5;
                const cy = Math.floor(idx / cols) + 0.5;
                const left = zn.x + (cx / cols) * zn.w;
                const top = zn.y + (cy / rows) * zn.h - 1;
                const col = String.fromCharCode(65 + (idx % cols));
                const row = Math.floor(idx / cols) + 1;
                return (
                  <div key={it.sku} className="zone-pin"
                    style={{ left: `${left}%`, top: `${top}%` }}>
                    <span className="pin-dot" style={{ background: zn.color }}/>
                    {it.sku} · {zn.id}-{String(row).padStart(2,'0')}{col}
                    <span className="pin-tail" />
                  </div>
                );
              });
            })()}
            {/* Aisle markers */}
            <div style={{ position: 'absolute', left: '50%', top: '74%', transform: 'translateX(-50%)', fontFamily: 'var(--mono)', fontSize: 10, color: 'var(--text-3)', letterSpacing: '0.2em' }}>—— MAIN AISLE ——</div>
            {/* Compass */}
            <div style={{ position: 'absolute', right: 10, top: 10, padding: '4px 8px', background: 'var(--bg)', border: '1px solid var(--border)', borderRadius: 6, fontSize: 11, fontFamily: 'var(--mono)', color: 'var(--text-2)' }}>N ↑ · 1:120</div>
          </div>
          <div className="legend" style={{ marginTop: 10 }}>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#2563eb' }} /> Apparel</div>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#0891b2' }} /> Electronics / Cold</div>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#16a34a' }} /> Home / Showroom</div>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#7c3aed' }} /> Receiving</div>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#d97706' }} /> Returns</div>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#dc2626' }} /> Outbound</div>
            <div className="legend-item"><span className="legend-swatch" style={{ background: '#5b6577' }} /> Bulk / Staging</div>
          </div>
        </div>

        <div className="card detail-panel" style={{ padding: 0 }}>
          <div className="card-head" style={{ borderBottom: '1px solid var(--border)' }}>
            <div>
              <div className="card-title">{z.name}</div>
              <div className="card-sub" style={{ marginTop: 2 }}>{z.type} · zone {z.id}</div>
            </div>
            <button className="icon-btn" style={{ marginLeft: 'auto' }}><Icon d={I.more} /></button>
          </div>
          <div className="card-body" style={{ paddingTop: 10 }}>
            <div className="kv"><span className="k">Capacity</span><span className="v">{z.items} / {z.capacity}</span></div>
            <div className="kv"><span className="k">Fill</span><span className="v">{Math.round(z.items / z.capacity * 100)}%</span></div>
            <div className="kv"><span className="k">Sales today</span><span className="v">${z.sales.toLocaleString()}</span></div>
            <div className="kv"><span className="k">Avg. dwell</span><span className="v">2d 4h</span></div>
            <div className="kv"><span className="k">Misplaced</span><span className="v">{z.id === 'A2' ? 3 : z.id === 'A4' ? 2 : 0}</span></div>
            <div className="kv"><span className="k">Last scan</span><span className="v" style={{ fontFamily: 'var(--mono)', fontSize: 12 }}>14:23:08</span></div>
            <div style={{ marginTop: 14, fontSize: 12, fontWeight: 600, color: 'var(--text-2)' }}>CONSTRAINTS</div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 8 }}>
              <span className="chip accent"><span className="dot" />Apparel only</span>
              <span className="chip"><span className="dot" />Max 480 SKUs</span>
              <span className="chip"><span className="dot" />Climate: ambient</span>
            </div>
          </div>
          <div className="card-foot">
            <button className="btn ghost">Open zone</button>
            <button className="btn primary">Assign task</button>
          </div>
        </div>
      </div>
    </>);

}

Object.assign(window, { Dashboard, MapPage, Stat, Sparkline, Heatmap, PriorityChip });