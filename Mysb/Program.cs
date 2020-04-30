using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Mysb.DataAccess;
using Mysb.Managers;
using TwoMQTT.Core;


namespace Mysb
{
    class Program : ConsoleProgram
    {
        static async Task Main(string[] args)
        {
            var p = new Program();
            p.MapOldEnvVariables();
            await p.ExecuteAsync(args);
        }

        protected override IServiceCollection ConfigureServices(HostBuilderContext hostContext, IServiceCollection services)
        {
            var sinkSect = hostContext.Configuration.GetSection(Models.SinkManager.Opts.Section);
            var sharedSect = hostContext.Configuration.GetSection(Models.Shared.Opts.Section);

            return services
                .Configure<Models.SinkManager.Opts>(sinkSect)
                .Configure<Models.Shared.Opts>(sharedSect)
                .AddSingleton<IFirmwareDAO, FirmwareDAO>()
                .AddSingleton<IHostedService, SinkManager>();
        }

        [Obsolete("Remove in the near future.")]
        private void MapOldEnvVariables()
        {
            var found = false;
            var foundOld = new List<string>();
            var mappings = new[]
            {
                new { Src = "MYSENSORS_SUBTOPIC", Dst = "MYSB__SUBTOPIC", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MYSENSORS_PUBTOPIC", Dst = "MYSB__PUBTOPIC", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MYSENSORS_AUTOID", Dst = "MYSB__AUTOIDENABLED", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MYSENSORS_NEXTID", Dst = "MYSB__NEXTID", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MYSENSORS_FIRMWAREBASEPATH", Dst = "MYSB__FIRMWAREBASEPATH", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MYSENSORS_NODES", Dst = "MYSB__RESOURCES", CanMap = false, Strip = "", Sep = "" },
                new { Src = "MQTT_TOPICPREFIX", Dst = "MYSB__MQTT__TOPICPREFIX", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MQTT_DISCOVERY", Dst = "MYSB__MQTT__DISCOVERYENABLED", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MQTT_DISCOVERYPREFIX", Dst = "MYSB__MQTT__DISCOVERYPREFIX", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MQTT_DISCOVERYNAME", Dst = "MYSB__MQTT__DISCOVERYNAME", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MQTT_BROKER", Dst = "MYSB__MQTT__BROKER", CanMap = true, Strip = "tcp://", Sep = "" },
                new { Src = "MQTT_USERNAME", Dst = "MYSB__MQTT__USERNAME", CanMap = true, Strip = "", Sep = "" },
                new { Src = "MQTT_PASSWORD", Dst = "MYSB__MQTT__PASSWORD", CanMap = true, Strip = "", Sep = "" },
            };

            foreach (var mapping in mappings)
            {
                var old = Environment.GetEnvironmentVariable(mapping.Src);
                if (string.IsNullOrEmpty(old))
                {
                    continue;
                }

                found = true;
                foundOld.Add($"{mapping.Src} => {mapping.Dst}");

                if (!mapping.CanMap)
                {
                    continue;
                }

                // Strip junk where possible
                if (!string.IsNullOrEmpty(mapping.Strip))
                {
                    old = old.Replace(mapping.Strip, string.Empty);
                }

                // Simple
                if (string.IsNullOrEmpty(mapping.Sep))
                {
                    Environment.SetEnvironmentVariable(mapping.Dst, old);
                }
            }


            if (found)
            {
                var loggerFactory = LoggerFactory.Create(builder => { builder.AddConsole(); });
                var logger = loggerFactory.CreateLogger<Program>();
                logger.LogWarning("Found old environment variables.");
                logger.LogWarning($"Please migrate to the new environment variables: {(string.Join(", ", foundOld))}");
                Thread.Sleep(5000);
            }
        }
    }
}
