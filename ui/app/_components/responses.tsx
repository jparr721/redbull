"use client";

import { useMemo, useState } from "react";
import { getLastCheckInTime, getResponses } from "@/queries/responses.query";
import { useQuery, useMutation } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import axios from "axios";
import { StickToBottom } from "use-stick-to-bottom";
import {
  Tool,
  ToolContent,
  ToolHeader,
  ToolOutput,
} from "@/components/ai-elements/tool";

export function Responses() {
  const [command, setCommand] = useState("");

  const { data: responses, isLoading } = useQuery({
    queryKey: ["responses"],
    queryFn: getResponses,
    refetchInterval: 1000,
  });

  const { data: checkInTime, isLoading: isLoadingCheckInTime } = useQuery({
    queryKey: ["checkInTime"],
    queryFn: getLastCheckInTime,
    refetchInterval: 1000,
  });

  const mutation = useMutation({
    mutationFn: (cmd: string) => axios.post("http://localhost:8000/command", { command: cmd }),
  });

  const currentDirectory = useMemo(() => {
    return responses?.at(responses.length - 1)?.currentDirectory ?? "";
  }, [responses]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!command.trim()) return;
    mutation.mutate(command);
    setCommand("");
  };


  if (isLoading) return <h1>Loading</h1>;

  return (
    <div className="flex flex-col h-full relative">
      <StickToBottom className="relative flex-1 overflow-y-hidden pb-24">
        <StickToBottom.Content className="flex flex-col gap-4 p-4">
          {responses?.map((response, i) => (
            <Tool key={response.id} defaultOpen={i == responses.length - 1}>
              <ToolHeader
                title={response.command}
                type={`tool-${response.command}`}
                state={response.stderr.trim().length > 0 ? "output-error" : "output-available"}
              />
              <ToolContent>
                <ToolOutput
                  output={response.stdout}
                  errorText={response.stderr.trim().length > 0 ? response.stderr : undefined}
                />
                <div className="flex flex-wrap gap-3 px-4 py-3 border-t border-border bg-muted/30">
                  {response.currentDirectory && (
                    <div className="flex items-center gap-1.5 text-xs">
                      <span className="text-muted-foreground">Directory:</span>
                      <span className="font-mono text-foreground">{response.currentDirectory}</span>
                    </div>
                  )}
                  <div className="flex items-center gap-1.5 text-xs">
                    <span className="text-muted-foreground">Executed:</span>
                    <span className="text-foreground">
                      {new Date(response.time).toLocaleString(undefined, {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                  </div>
                </div>
              </ToolContent>
            </Tool>
          ))}
        </StickToBottom.Content>
      </StickToBottom>

      {/* Fixed bottom input container */}
      <div className="fixed bottom-0 left-0 right-0 bg-background border-t border-border p-4 z-10">
        {currentDirectory && (
          <div>
            <div className="mb-2 text-sm text-muted-foreground">
              Current directory: <span className="font-mono text-foreground">{currentDirectory}</span>
            </div>
            <div className="mb-2 text-sm text-muted-foreground">
              {
              isLoadingCheckInTime
                ? <span className="font-mono text-foreground">Loading</span>
                : <span>
                    Last Checkin: <span className="font-mono text-foreground">
                      {`${checkInTime?.checkInTime ?? "No check in yet"}ms ago`}
                    </span>
                  </span>
              }
            </div>
          </div>
        )}
        <form onSubmit={handleSubmit}>
          <Input
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            placeholder="Enter command..."
            className="w-full"
          />
        </form>
      </div>
    </div>
  );
}
