import * as TooltipPrimitive from "@radix-ui/react-tooltip";
import * as React from "react";

import { cn } from "@/lib/utils";

const TooltipProvider = TooltipPrimitive.Provider;

const Tooltip = TooltipPrimitive.Root;

const TooltipTrigger = TooltipPrimitive.Trigger;

const TooltipContent = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
  <TooltipPrimitive.Portal>
    <TooltipPrimitive.Content
      ref={ref}
      sideOffset={sideOffset}
      className={cn(
        "z-50 origin-[--radix-tooltip-content-transform-origin] overflow-hidden rounded-md bg-primary px-3 py-1.5 text-xs text-primary-foreground animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
        className,
      )}
      {...props}
    />
  </TooltipPrimitive.Portal>
));
TooltipContent.displayName = TooltipPrimitive.Content.displayName;

interface PassthroughTooltipProps {
  triggerProps?: TooltipPrimitive.TooltipTriggerProps;
  contentProps?: TooltipPrimitive.TooltipContentProps;
}

const CopyTooltip = ({
  copyData,
  children,
  triggerProps,
  contentProps,
}: React.PropsWithChildren<{
  copyData: (ClipboardItem | string)[];
}> &
  PassthroughTooltipProps) => {
  const TOOLTIP_DEFAULT = "Sao chép";
  const TOOLTIP_SUCCESS = "Đã sao chép";

  const [tooltip, setTooltip] = React.useState(TOOLTIP_DEFAULT);
  const ref = React.useRef<HTMLButtonElement>(null);

  const [open, setOpen] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  const isSuccessTooltip = tooltip === TOOLTIP_SUCCESS;

  const handleClick = React.useCallback(() => {
    setCopied(true);

    navigator.clipboard
      .write(
        copyData.map((data) =>
          typeof data === "string"
            ? new ClipboardItem({ "text/plain": data })
            : data,
        ),
      )
      .then(() => {
        setTooltip(TOOLTIP_SUCCESS);

        setTimeout(() => {
          setCopied(false);

          // Use setTimeout to allow animation to finish
          setTimeout(() => setTooltip(TOOLTIP_DEFAULT), 100);
        }, 1000);
      });
  }, []);

  return (
    <Tooltip
      open={open || copied}
      onOpenChange={setOpen}
      disableHoverableContent={true}
    >
      <TooltipTrigger
        ref={ref}
        onClick={handleClick}
        {...triggerProps}
        className={cn("hover:cursor-pointer", triggerProps?.className)}
      >
        {children}
      </TooltipTrigger>
      <TooltipContent
        {...contentProps}
        className={cn(
          isSuccessTooltip && "bg-green-900",
          contentProps?.className,
        )}
      >
        {tooltip}
      </TooltipContent>
    </Tooltip>
  );
};

const CopyableSpan = ({
  text,
  triggerProps,
  contentProps,
  ...props
}: React.HTMLAttributes<HTMLSpanElement> &
  PassthroughTooltipProps & { text: any }) =>
  text != null ? (
    <CopyTooltip
      copyData={[String(text)]}
      triggerProps={{ asChild: true, ...triggerProps }}
      contentProps={contentProps}
    >
      <span {...props}>{text}</span>
    </CopyTooltip>
  ) : (
    <span {...props}></span>
  );

export {
  CopyableSpan,
  CopyTooltip,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
};
