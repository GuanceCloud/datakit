#encoding: utf-8

from datakit_framework import DataKitFramework

class Demo(DataKitFramework):
    name = 'Demo'
    interval = 10 # triggered interval seconds.

    # if your datakit ip is 127.0.0.1 and port is 9529, you won't need use this,
    # just comment it.
    # def __init__(self, **kwargs):
    #     super().__init__(ip = '127.0.0.1', port = 9529)

    # General report example.
    def run(self):
        print("Demo")
        data = [
                {
                    "measurement": "abc",
                    "tags": {
                    "t1": "b",
                    "t2": "d"
                    },
                    "fields": {
                    "f1": 123,
                    "f2": 3.4,
                    "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                },

                {
                    "measurement": "def",
                    "tags": {
                    "t1": "b",
                    "t2": "d"
                    },
                    "fields": {
                    "f1": 123,
                    "f2": 3.4,
                    "f3": "strval"
                    },
                    # "time": 1624550216 # you don't need this
                }
            ]

        in_data = {
            'M':data, # 'M' for metrics, 'L' for logging, 'R' for rum, 'O' for object, 'CO' for custom object, 'E' for event.
            'input': "datakitpy"
        }

        return self.report(in_data) # you must call self.report here

    # # KeyEvent report example.
    # def run(self):
    #     print("Demo")

    #     tags = {"tag1": "val1", "tag2": "val2"}
    #     date_range = 10
    #     status = 'info'
    #     event_id = 'event_id'
    #     title = 'title'
    #     message = 'message'
    #     kwargs = {"custom_key1":"custom_value1", "custom_key2": "custom_value2", "custom_key3": "custom_value3"}

    #     # Feed df_source=user event.
    #     user_id="user_id"
    #     return self.feed_user_event(
    #         user_id,
    #         tags, date_range, status, event_id, title, message, **kwargs
    #         )

    #     # Feed df_source=monitor event.
    #     dimension_tags='{"host":"web01"}' # dimension_tags must be the String(JSON format).
    #     return self.feed_monitor_event(
    #         dimension_tags,
    #         tags, date_range, status, event_id, title, message, **kwargs
    #         )

    #     # Feed df_source=system event.
    #     return self.feed_system_event(
    #         tags, date_range, status, event_id, title, message, **kwargs
    #         )

    # # metrics, logging, object example.
    # def run(self):
    #     print("Demo")

    #     measurement = "mydata"
    #     tags = {"tag1": "val1", "tag2": "val2"}
    #     fields = {"custom_field1": "val1","custom_field2": 1000}
    #     kwargs = {"custom_key1":"custom_value1", "custom_key2": "custom_value2", "custom_key3": "custom_value3"}

    #     # Feed metrics example.
    #     return self.feed_metric(
    #         measurement=measurement,
    #         tags=tags,
    #         fields=fields,
    #         **kwargs
    #         )

    #     # Feed logging example.
    #     message = "This is the message for testing"
    #     return self.feed_logging(
    #         source=measurement,
    #         tags=tags,
    #         message=message,
    #         **kwargs
    #         )

    #     # Feed object example.
    #     name = "name"
    #     return self.feed_object(
    #         cls=measurement,
    #         name=name,
    #         tags=tags,
    #         fields=fields,
    #         **kwargs
    #         )