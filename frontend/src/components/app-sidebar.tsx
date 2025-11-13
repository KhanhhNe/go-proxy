import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { PAGES, usePageStore } from "@/state";
import { Globe, Home } from "lucide-react";

const items = [
  {
    page: PAGES.index,
    icon: Home,
    name: "Trang chủ",
  },
  {
    page: PAGES.servers,
    icon: Globe,
    name: "Nguồn proxy",
  },
];

export function AppSidebar() {
  const page = usePageStore((state) => state.page);
  const changePage = usePageStore((state) => state.setPage);

  return (
    <Sidebar collapsible="icon">
      <SidebarTrigger className="ml-auto" />
      <SidebarHeader />
      <SidebarContent>
        <SidebarGroup>
          {items.map((item) => (
            <SidebarMenuItem>
              <SidebarMenuButton asChild isActive={page === item.page}>
                <a href="#" onClick={() => changePage(item.page)}>
                  <item.icon /> <span>{item.name}</span>
                </a>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter />
    </Sidebar>
  );
}
