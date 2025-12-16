"use client";

import { useState } from "react";
import { getResponses } from "@/queries/responses.query";
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
  const [currentDirectory, setCurrentDirectory] = useState("");

  const { data: responses, isLoading } = useQuery({
    queryKey: ["responses"],
    queryFn: getResponses,
    refetchInterval: 1000,
  });

  const mutation = useMutation({
    mutationFn: (cmd: string) => axios.post("http://localhost:8000/command", { command: cmd }),
  });

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
              </ToolContent>
            </Tool>
          ))}
        </StickToBottom.Content>
      </StickToBottom>

      {/* Fixed bottom input container */}
      <div className="fixed bottom-0 left-0 right-0 bg-background border-t border-border p-4 z-10">
        {currentDirectory && (
          <div className="mb-2 text-sm text-muted-foreground">
            Current directory: <span className="font-mono text-foreground">{currentDirectory}</span>
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
