#!/usr/bin/env ruby

require 'yaml'

request = YAML.load_file('/input/object.yaml')
instance = YAML.load_file('/tmp/transfer/redis-instance.yaml')

requested_name = request.dig('spec', 'name') || raise('spec.name cannot be empty')
instance['metadata']['name'] = requested_name

File.open('/output/instance.yaml', 'w') do |f|
  YAML.dump(instance, f)
end
