"use client";

import { ReactNode, useMemo } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "./ui/sidebar";
import { AppSidebar } from "./app-sidebar";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "./ui/breadcrumb";
import { Separator } from "./ui/separator";

interface Props {
  children: ReactNode;
}

function formatSegment(segment: string): string {
  return segment
    .split("-")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export function LayoutWrapper({ children }: Props) {
  const pathname = usePathname();

  const breadcrumbs = useMemo(() => {
    const segments = pathname.split("/").filter(Boolean);

    if (segments.length === 0) {
      return [
        <BreadcrumbItem key="home" className="hidden md:block">
          <BreadcrumbPage>Home</BreadcrumbPage>
        </BreadcrumbItem>,
      ];
    }

    const items = [
      <BreadcrumbItem key="home" className="hidden md:block">
        <BreadcrumbLink asChild>
          <Link href="/">Home</Link>
        </BreadcrumbLink>
      </BreadcrumbItem>,
    ];

    segments.forEach((segment, index) => {
      const isLast = index === segments.length - 1;
      const href = "/" + segments.slice(0, index + 1).join("/");
      const displayName = formatSegment(segment);

      items.push(
        <BreadcrumbSeparator key={`separator-${index}`} />
      );

      if (isLast) {
        items.push(
          <BreadcrumbItem key={segment} className="hidden md:block">
            <BreadcrumbPage>{displayName}</BreadcrumbPage>
          </BreadcrumbItem>
        );
      } else {
        items.push(
          <BreadcrumbItem key={segment} className="hidden md:block">
            <BreadcrumbLink asChild>
              <Link href={href}>{displayName}</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
        );
      }
    });

    return items;
  }, [pathname]);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              {breadcrumbs}
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col overflow-hidden">
          <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6 p-4 h-full overflow-hidden">
            {children}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
