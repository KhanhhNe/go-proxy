import { Events } from "@wailsio/runtime";
import { useEffect } from "react";
import { AppSidebar } from "./components/app-sidebar";
import { SidebarProvider } from "./components/ui/sidebar";
import { debounce } from "./lib/utils";
import { PageIndex } from "./pages";
import { PageServers } from "./pages/servers";
import {
  PAGES,
  useAppStateStore,
  useManagerStore,
  usePageStore,
} from "./state";

function App() {
  const page = usePageStore((state) => state.page);
  const fetchManager = debounce(
    useManagerStore((state) => state.fetchManager),
    500,
  );
  const fetchAppState = useAppStateStore((state) => state.fetchState);

  useEffect(() => {
    fetchManager();
    fetchAppState();
    // const appStateInt = setInterval(fetchAppState, 1000);

    Events.On("goproxy:data-changed", () => setTimeout(fetchManager));

    return () => {
      // clearInterval(appStateInt);
      Events.Off("goproxy:data-changed");
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <SidebarProvider>
      <AppSidebar />
      <main id="App" className="m-5 w-full">
        {page === PAGES.index && <PageIndex />}
        {page === PAGES.servers && <PageServers />}
      </main>
    </SidebarProvider>
  );
}

export default App;
