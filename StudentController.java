package com.example.rdemo;

import lombok.Builder;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
import org.joda.time.DateTime;
import org.springframework.messaging.handler.annotation.DestinationVariable;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.stereotype.Controller;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

/**
 * Description:
 *
 * @author Jeffsky
 * @date 2020-04-28
 */
@Controller
@Slf4j
public class StudentController {

  @MessageMapping("student.v1.upsert")
  public Mono<Result<Student>> upsertStudent(@Payload Student student) {
    log.info("upsert student: {}", student);
    student.setId(System.currentTimeMillis());
    return Mono.just(Result.<Student>builder().data(student).build());
  }

  @MessageMapping("student.v1.noop.{txt}")
  public Mono<Void> forNoop(@DestinationVariable("txt") String txt) {
    log.info("---> noop: {}", txt);
    return Mono.empty();
  }

  @MessageMapping("student.v1.{id}")
  public Mono<Student> getStudent(@DestinationVariable("id") Long id) {
    return Mono.just(Student.builder().id(id).name("foobar").birth("2020").build());
  }

  @MessageMapping("students.v1")
  public Flux<Student> listStudents() {
    return Flux.range(0, 10)
        .map(
            n ->
                Student.builder()
                    .id(Long.valueOf(n))
                    .birth(DateTime.now().toString("yyyy-MM-dd"))
                    .name("Foobar" + n)
                    .build());
  }

  @Data
  @Builder
  public static final class Result<T> {
    private int code;
    private String message;
    private T data;
  }

  @Data
  @Builder
  public static final class Student {
    private Long id;
    private String name;
    private String birth;
  }
}
