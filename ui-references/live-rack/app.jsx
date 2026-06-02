/* global React, ReactDOM */
const { useState, useEffect } = React;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "theme": "light",
  "accent": "#2563eb",
  "density": "Balanced",
  "showHints": true
}/*EDITMODE-END*/;

function App() {
  const [tweaks, setTweak] = window.useTweaks(TWEAK_DEFAULTS);

  const [page, setPage] = useState('dashboard');
  const [role, setRole] = useState({ id: 'manager', name: 'Avery Chen', title: 'Ops Manager', initials: 'AC' });
  const [drawer, setDrawer] = useState(false);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', tweaks.theme);
    document.documentElement.style.setProperty('--accent', tweaks.accent);
    document.documentElement.style.setProperty('--accent-soft',
      `color-mix(in oklab, ${tweaks.accent} 18%, transparent)`);
    document.documentElement.style.setProperty('--accent-text', tweaks.accent);
  }, [tweaks.theme, tweaks.accent]);

  const screens = {
    dashboard:    <window.Dashboard accent={tweaks.accent} />,
    map:          <window.MapPage accent={tweaks.accent} />,
    scanner:      <window.Scanner accent={tweaks.accent} />,
    inventory:    <window.Inventory />,
    tasks:        <window.Tasks />,
    pipelines:    <window.Pipelines />,
    analytics:    <window.Analytics accent={tweaks.accent} />,
    integrations: <window.Integrations />,
    users:        <window.Users />,
  };

  return (
    <div className="app" data-screen-label={`live-rack · ${page}`} data-drawer={drawer ? 'open' : 'closed'}>
      <div className="mobile-drawer-backdrop" onClick={()=>setDrawer(false)} />
      <window.Sidebar page={page} setPage={(p)=>{setPage(p); setDrawer(false);}} role={role} accent={tweaks.accent} />
      <div className="main">
        <window.Topbar page={page} role={role} setRole={setRole}
          density={tweaks.density} setDensity={(d)=>setTweak('density', d)} />
        <window.MobileTopbar page={page} openDrawer={()=>setDrawer(true)} />
        <div className="content">
          {screens[page]}
        </div>
        <window.MobileTabbar page={page} setPage={setPage} />
      </div>

      {window.TweaksPanel && (
        <window.TweaksPanel>
          <window.TweakSection title="Theme">
            <window.TweakRadio label="Mode" value={tweaks.theme} onChange={(v)=>setTweak('theme', v)}
              options={[{value:'light',label:'Light'},{value:'dark',label:'Dark'}]} />
            <window.TweakColor label="Accent" value={tweaks.accent} onChange={(v)=>setTweak('accent', v)}
              options={['#2563eb','#16a34a','#7c3aed','#d97706']} />
          </window.TweakSection>
          <window.TweakSection title="Layout">
            <window.TweakRadio label="Density" value={tweaks.density} onChange={(v)=>setTweak('density', v)}
              options={[{value:'Compact',label:'Compact'},{value:'Balanced',label:'Balanced'},{value:'Roomy',label:'Roomy'}]} />
            <window.TweakToggle label="Inline hints" value={tweaks.showHints} onChange={(v)=>setTweak('showHints', v)} />
          </window.TweakSection>
          <window.TweakSection title="Jump to screen">
            <window.TweakSelect label="Screen" value={page} onChange={setPage}
              options={[
                {value:'dashboard',label:'Overview'},
                {value:'map',label:'Map & Zones'},
                {value:'scanner',label:'Scanner'},
                {value:'inventory',label:'Inventory'},
                {value:'tasks',label:'Tasks'},
                {value:'pipelines',label:'Pipelines'},
                {value:'analytics',label:'Analytics'},
                {value:'integrations',label:'Integrations'},
                {value:'users',label:'Users & Access'},
              ]} />
          </window.TweakSection>
        </window.TweaksPanel>
      )}
    </div>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />);
