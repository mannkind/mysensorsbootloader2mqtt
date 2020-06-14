using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using MQTTnet;
using MQTTnet.Extensions.ManagedClient;
using Mysb.DataAccess;
using Mysb.Models.Options;
using TwoMQTT.Core.Interfaces;
using TwoMQTT.Core.Models;
using TwoMQTT.Core.Utils;

namespace Mysb.Liasons
{
    /// <summary>
    /// An class representing a managed way to interact with a sink.
    /// </summary>
    public class MQTTLiason : IMQTTLiason<object, object>
    {
        /// <summary>
        /// Initializes a new instance of the MQTTLiason class.
        /// </summary>
        /// <param name="logger"></param>
        /// <param name="generator"></param>
        /// <param name="sharedOpts"></param>
        public MQTTLiason(ILogger<MQTTLiason> logger, IMQTTGenerator generator, IFirmwareDAO loader, IOptions<SharedOpts> sharedOpts)
        {
            this.Logger = logger;
            this.Generator = generator;
            this.Loader = loader;
            this.AutoIDEnabled = sharedOpts.Value.AutoIDEnabled;
            this.NextID = sharedOpts.Value.NextID;
            this.SubTopic = sharedOpts.Value.SubTopic;
            this.PubTopic = sharedOpts.Value.PubTopic;

            this.Logger.LogInformation(
                $"FirmwareBasePath: {sharedOpts.Value.FirmwareBasePath}\n" +
                $"AutoIDEnabled: {sharedOpts.Value.AutoIDEnabled}\n" +
                $"NextID: {sharedOpts.Value.NextID}\n" +
                $"SubTopic: {sharedOpts.Value.SubTopic}\n" +
                $"PubTopic: {sharedOpts.Value.PubTopic}\n" +
                $"Resources: {string.Join("; ", sharedOpts.Value.Resources)}\n" +
                $""
            );
        }

        /// <inheritdoc />
        public IEnumerable<(string topic, string payload)> MapData(object input) =>
            new List<(string, string)>();

        /// <inheritdoc />
        public async Task HandleCommand(IManagedMqttClient client, string topic, string payload,
            CancellationToken cancellationToken = default)
        {
            var results = new List<object>();

            var bootloaderCommand = Const.FirmwareBootloaderCommandTopic.Replace("/+/+", string.Empty);
            if (topic.StartsWith(bootloaderCommand))
            {
                this.HandleBootloaderCommand(topic, payload, cancellationToken);
                return;
            }

            var parts = topic.Replace($"{this.SubTopic}/", string.Empty).Split('/');
            if (parts.Length != 5)
            {
                this.Logger.LogError("Unable to determine the nodeId from the topic; aborting.");
                return;
            }

            var nodeId = parts[0];
            var idRequest = $"{this.SubTopic}/{Const.IdRequestTopicPartial}";
            var firmwareConfigRequest = $"{this.SubTopic}/{nodeId}/{Const.FirmwareConfigRequestTopicPartial}".Replace("+/", string.Empty);
            var firmwareRequest = $"{this.SubTopic}/{nodeId}/{Const.FirmwareRequestTopicPartial}".Replace("+/", string.Empty);

            switch (topic)
            {
                case string s when s == idRequest:
                    await this.HandleIdRequest(client, cancellationToken);
                    break;

                case string s when s == firmwareConfigRequest:
                    await this.HandleFirmwareConfigRequest(client, nodeId, payload, cancellationToken);
                    break;

                case string s when s == firmwareRequest:
                    await this.HandleFirmwareRequest(client, nodeId, payload, cancellationToken);
                    break;

                case string s when s.StartsWith(bootloaderCommand):
                    break;
            }
        }


        /// <inheritdoc />
        public IEnumerable<string> Subscriptions()
        {
            var topics = new List<string>
            {
                $"{this.SubTopic}/{Const.IdRequestTopicPartial}",
                $"{this.SubTopic}/{Const.FirmwareConfigRequestTopicPartial}",
                $"{this.SubTopic}/{Const.FirmwareRequestTopicPartial}",
                Const.FirmwareBootloaderCommandTopic,
            };

            return topics;
        }

        /// <inheritdoc />
        public IEnumerable<(string slug, string sensor, string type, MQTTDiscovery discovery)> Discoveries() =>
            new List<(string slug, string sensor, string type, MQTTDiscovery discovery)>();

        /// <summary>
        /// The logger used internally.
        /// </summary>
        private readonly ILogger<MQTTLiason> Logger;

        /// <summary>
        /// The MQTT generator used for things such as availability topic, state topic, command topic, etc.
        /// </summary>
        private IMQTTGenerator Generator;

        /// <summary>
        /// The firmware loader that loads firmware from disk.
        /// </summary>
        private readonly IFirmwareDAO Loader;

        /// <summary>
        /// 
        /// </summary>
        private ushort NextID;

        /// <summary>
        /// 
        /// </summary>
        private readonly bool AutoIDEnabled;

        /// <summary>
        /// 
        /// </summary>
        private readonly string SubTopic;

        /// <summary>
        /// 
        /// </summary>
        private readonly string PubTopic;

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="string"></typeparam>
        /// <typeparam name="FirmwareConfigReqResp"></typeparam>
        /// <returns></returns>
        private readonly ConcurrentDictionary<string, string> BootloaderCommands = new ConcurrentDictionary<string, string>();

        /// <summary>
        /// 
        /// </summary>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task HandleIdRequest(IManagedMqttClient client, CancellationToken cancellationToken = default)
        {
            if (!this.AutoIDEnabled)
            {
                return;
            }

            this.NextID += 1;

            var respTopic = $"{this.PubTopic}/{Const.IdResponseTopicPartial}";
            var respPayload = this.NextID.ToString();
            await this.PublishAsync(client, respTopic, respPayload, cancellationToken);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task HandleFirmwareConfigRequest(IManagedMqttClient client, string nodeId, string payload,
            CancellationToken cancellationToken = default)
        {
            var respTopic = $"{this.PubTopic}/{nodeId}/{Const.FirmwareConfigResponseTopicPartial}";

            if (this.BootloaderCommands.ContainsKey(nodeId) && this.BootloaderCommands.TryRemove(nodeId, out var bootloaderPayload))
            {
                await this.PublishAsync(client, respTopic, bootloaderPayload, cancellationToken);
                return;
            }

            var respPayload = await this.Loader.FirmwareConfigAsync(nodeId, payload, cancellationToken);
            await this.PublishAsync(client, respTopic, respPayload, cancellationToken);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private async Task HandleFirmwareRequest(IManagedMqttClient client, string nodeId, string payload,
            CancellationToken cancellationToken = default)
        {
            var respTopic = $"{this.PubTopic}/{nodeId}/{Const.FirmwareResponseTopicPartial}";
            var respPayload = await this.Loader.FirmwareAsync(nodeId, payload, cancellationToken);
            await this.PublishAsync(client, respTopic, respPayload, cancellationToken);
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="topic"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        private void HandleBootloaderCommand(string topic, string payload,
            CancellationToken cancellationToken = default)
        {
            var (nodeId, resp) = this.Loader.BootloaderCommand(topic, payload);
            if (string.IsNullOrEmpty(nodeId))
            {
                return;
            }

            this.BootloaderCommands[nodeId] = resp;
            return;
        }


        /// <summary>
        /// Publish topics + payloads
        /// </summary>
        private async Task PublishAsync(IManagedMqttClient client, string topic, string payload,
            CancellationToken cancellationToken = default)
        {
            this.Logger.LogDebug($"Publishing '{payload}' on '{topic}'");
            await client.PublishAsync(
                new MqttApplicationMessageBuilder()
                    .WithTopic(topic)
                    .WithPayload(payload)
                    .Build(),
                cancellationToken
            );
        }
    }
}