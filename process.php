<?php

use Swaggest\JsonSchema\Schema;
use Swaggest\JsonSchema\Exception as SwaggestException;

require_once __DIR__ . '/vendor/autoload.php';

const COMMON = 'common';

$schema = Schema::import(json_decode(file_get_contents(__DIR__ . '/schema.json')));
try {
    $data = json_decode(file_get_contents(__DIR__ . '/source.json'));
    $schema->in($data);

    $ids = [];
    $now = new DateTimeImmutable();
    $results = [];
    foreach ($data as $key => &$item) {
        $item = (array) $item;
        $validSince = new DateTimeImmutable($item['validSince']);
        $validUntil = new DateTimeImmutable($item['validUntil']);

        $item['validSince'] = $validSince->setTimezone(new DateTimeZone('UTC'));
        $item['validUntil'] = $validUntil->setTimezone(new DateTimeZone('UTC'));

        if (isset($ids[$item['id']])) {
            echo "The ID '{$item['id']}' already exists.", PHP_EOL;
            exit(1);
        }
        $ids[$item['id']] = true;

        if ($now < $validSince || $now > $validUntil) {
            continue;
        }

        $item['limitedTo'] ??= [COMMON];
        foreach ($item['limitedTo'] as $limitedTo) {
            $results[$limitedTo] ??= [];
            $results[$limitedTo][] = $item;
        }
    }

    $clients = json_decode(file_get_contents(__DIR__ . '/clients.json'));
    assert(is_array($clients));
    foreach ($clients as $client) {
        assert(is_string($client));
        $results[$client] ??= [];
    }

    $results[COMMON] ??= [];
    foreach ($results as $key => $items) {
        if ($key === COMMON) {
            continue;
        }

        $results[$key] = [...$items, ...$results[COMMON]];
    }

    $results = array_map(function (array $items) {
        usort($items, fn (array $a, array $b) => -($a['validSince'] <=> $b['validSince']));
        return array_map(function (array $item) {
            $item['validSince'] = $item['validSince']->format('c');
            $item['validUntil'] = $item['validUntil']->format('c');
            unset($item['limitedTo']);

            return $item;
        }, $items);
    }, $results);

    foreach ($results as $limitedTo => $result) {
        $filename = "results.{$limitedTo}.json";
        $filenameMin = "results.{$limitedTo}.min.json";

        file_put_contents(__DIR__ . "/{$filename}", json_encode($result, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));
        file_put_contents(__DIR__ . "/{$filenameMin}", json_encode($result, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));

        $schema->in(json_decode(file_get_contents(__DIR__ . "/{$filename}")));
        $schema->in(json_decode(file_get_contents(__DIR__ . "/{$filenameMin}")));
    }
} catch (SwaggestException $e) {
    echo $e->getMessage(), PHP_EOL;
    exit(1);
}
